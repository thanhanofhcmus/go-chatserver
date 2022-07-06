package main

import (
	"log"
	"net/http"

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
	gClientRemover = make(chan *Client)
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./frontend/dist")))
	http.HandleFunc("/connect", handleConnections)

	go StartRemoveClient()

	go GetRedisClient().StartSendConvList()
	go GetRedisClient().StartListening()

	log.Printf("%s is serving", gServerId)
	http.ListenAndServe(":8000", nil)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	socket, err := gUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer socket.Close()
	client := NewClient(socket)

	// TODO: push to redis
	if err := socket.WriteJSON(IdMessage{Id: client.Id, Type: ID_ACTION}); err != nil {
		log.Println("Write IdMessage to client error: ", err)
		return
	}
	GetRedisClient().SendMessage(NewServerRequestMessage(CLIENT_CONNECTED_ACTION, client.Id))

	gClients.Store(client.Id, client)
	conv := NewPeerConv(&client)
	gConvs.Store(client.Id, conv)

	go client.StartWrite()
	// start here instead of spawn new goroutine so that we can defer socket.close()
	client.StartRead()
}

func gRemoveClient(client *Client) {
	go func() {
		gClientRemover <- client
	}()
}

func StartRemoveClient() {
	for client := range gClientRemover {
		log.Println("Remove client", client)
		gClients.Delete(client.Id)
		gConvs.Delete(client.Id)
		GetRedisClient().SendMessage(NewServerRequestMessage(CLIENT_DISCONNECTED_ACTION, client.Id))
		gConvs.Range(func(_ string, conv Conv) bool {
			if groupConv, ok := conv.(*GroupConv); ok {
				groupConv.RemoveClient(client)
			}
			return true
		})
	}
}
