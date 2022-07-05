package main

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Conv interface {
	Id() string
	DeliverTextMessage(TextMessage)
}

type PeerConv struct {
	client *Client
}

func (c PeerConv) Id() string {
	return c.client.Id
}

func (c PeerConv) DeliverTextMessage(msg TextMessage) {
	c.client.SendTextMessage(msg)
}

func (c PeerConv) MarshalJSON() ([]byte, error) {
	return json.Marshal((&struct {
		Id   string `json:"id"`
		Type string `json:"type"`
	}{Id: c.client.Id, Type: "peer"}))
}

func NewPeerConv(client *Client) PeerConv {
	return PeerConv{client: client}
}

type GroupConv struct {
	clients       concurrentMap[string, *Client]
	id            string
	clientRemover chan *Client
}

func (c *GroupConv) Id() string {
	return c.id
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
		Id   string `json:"id"`
		Type string `json:"type"`
	}{Id: c.id, Type: "group"}))
}

func NewGroupConv(clients ...*Client) *GroupConv {
	clientMap := make(map[*Client]bool)
	for _, client := range clients {
		clientMap[client] = true
	}
	conv := &GroupConv{
		clients:       NewConcurrentMap[string, *Client](),
		id:            uuid.NewString(),
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
