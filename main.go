package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
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
	rand.Seed(time.Now().UnixNano())

	http.Handle("/", http.FileServer(http.Dir("./frontend/dist")))
	http.HandleFunc("/connect", handleConnections)

	go removeClient()

	fmt.Println("serving")
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

	if err := socket.WriteJSON(IdMessage{Id: client.Id, Type: "id"}); err != nil {
		log.Println("Write IsMessage to client error: ", err)
		return
	}

	gClients.Store(client.Id, client)
	conv := NewPeerConv(&client)
	gConvs.Store(client.Id, conv)

	go client.StartWrite()
	client.StartRead() // start here instead of spawn new goroutine so that we can defer socket.close()
}

func removeClient() {
	for client := range gClientRemover {
		log.Println("removeClient", client)
		gClients.Delete(client.Id)
		gConvs.Delete(client.Id)
	}
}
