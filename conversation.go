package main

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
)

type Conv interface {
	Id() string
	ServerId() string
	DeliverTextMessage(TextMessage)
}

type PeerConv struct {
	client   *Client
	serverId string
}

func (c PeerConv) Id() string {
	return c.client.Id
}

func (c PeerConv) ServerId() string {
	return c.serverId
}

func (c PeerConv) DeliverTextMessage(msg TextMessage) {
	c.client.SendTextMessage(msg)
}

func (c PeerConv) MarshalJSON() ([]byte, error) {
	return json.Marshal((&struct {
		Id       string `json:"id"`
		Type     string `json:"type"`
		ServerId string `json:"severId"`
	}{Id: c.client.Id, ServerId: c.serverId, Type: "peer"}))
}

func NewPeerConv(client *Client) PeerConv {
	return PeerConv{client: client, serverId: gServerId}
}

type GroupConv struct {
	clients       concurrentMap[string, *Client]
	id            string
	serverId      string
	clientRemover chan *Client
}

func (c *GroupConv) Id() string {
	return c.id
}

func (c *GroupConv) ServerId() string {
	return c.serverId
}

func (c *GroupConv) AddClient(client *Client) {
	c.clients.Store(client.Id, client)
}

func (c *GroupConv) RemoveClient(client *Client) {
	go func() {
		c.clientRemover <- client
	}()
}

func (c *GroupConv) DeliverTextMessage(msg TextMessage) {
	c.clients.RRange(func(_ string, client *Client) bool {
		if client.Id != msg.SenderId {
			newMessage := TextMessage{
				SenderId:   msg.SenderId,
				ReceiverId: client.Id,
				Message:    msg.Message,
				Type:       TEXT_ACTION,
			}
			client.SendTextMessage(newMessage)
		}
		return true
	})
}

func (c *GroupConv) MarshalJSON() ([]byte, error) {
	return json.Marshal((&struct {
		Id       string `json:"id"`
		Type     string `json:"type"`
		ServerId string `json:"serverId"`
	}{Id: c.id, ServerId: c.serverId, Type: "group"}))
}

func NewGroupConv(clients ...*Client) *GroupConv {
	clientMap := make(map[*Client]bool)
	for _, client := range clients {
		clientMap[client] = true
	}
	conv := &GroupConv{
		clients:       NewConcurrentMap[string, *Client](),
		id:            uuid.NewString(),
		serverId:      gServerId,
		clientRemover: make(chan *Client),
	}
	// start this group remove client goroutine
	go func() {
		for client := range conv.clientRemover {
			conv.clients.Delete(client.Id)
		}
	}()
	return conv
}

type RemoteConv struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	ServerID string `json:"serverId"`
}

func NewRemoteConvFromJSON(source string) (c RemoteConv, err error) {
	err = json.Unmarshal([]byte(source), &c)
	return
}

func (c RemoteConv) Id() string {
	return c.ID
}

func (c RemoteConv) ServerId() string {
	return c.ServerID
}

func (c RemoteConv) DeliverTextMessage(msg TextMessage) {
	log.Printf("Send message %s to client %s\n", msg.Message, c.ID)
}
