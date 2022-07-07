package main

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	gServerId = uuid.NewString()
	gClients  = NewConcurrentMap[string, Client]()
	gConvs    = NewConcurrentMap[string, Conv]()
	gUpgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	gClientRemover = make(chan string)
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./frontend/dist")))
	http.HandleFunc("/connect", handleConnections)

	go StartRemoveClient()

	go GetRedisClient().StartSendConvList()
	go GetRedisClient().StartListening()

	time.Sleep(time.Second)

	log.Printf("%s is serving\n", gServerId)
	http.ListenAndServe(":8080", nil)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	socket, err := gUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer socket.Close()
	client := NewClient(socket)

	if err := socket.WriteJSON(IdMessage{Id: client.Id, Type: ID_ACTION}); err != nil {
		log.Println("Write IdMessage to client error: ", err)
		return
	}

	GetRedisClient().SendMessage(NewServerRequestMessage(
		CLIENT_CONNECTED_ACTION,
		ClientConnectedMessage{
			Id:       client.Id,
			ServerId: gServerId,
		},
	))

	gClients.Store(client.Id, client)
	conv := NewPeerConv(&client)
	gConvs.Store(client.Id, conv)

	go client.StartWrite()
	// start here instead of spawn new goroutine so that we can defer socket.close()
	client.StartRead()
}

func gRemoveClient(clientId string) {
	go func() {
		GetRedisClient().SendMessage(NewServerRequestMessage(CLIENT_DISCONNECTED_ACTION, clientId))
		gClientRemover <- clientId
	}()
}

func StartRemoveClient() {
	for clientId := range gClientRemover {
		log.Println("Remove client", clientId)
		gClients.Delete(clientId)
		gConvs.Delete(clientId)
		gConvs.Range(func(_ string, conv Conv) bool {
			if groupConv, ok := conv.(*GroupConv); ok {
				groupConv.RemoveClient(clientId)
			}
			return true
		})
	}
}
