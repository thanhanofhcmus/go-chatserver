package main

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Conversation interface {
	Id() string
	DeliverMessage(TextMessage)
}

type PeerConversation struct {
	client *Client
}

func (c PeerConversation) Id() string {
	return c.client.Id
}

func (c PeerConversation) DeliverMessage(msg TextMessage) {
	c.client.SendTextMessage(msg)
}

func (c PeerConversation) MarshalJSON() ([]byte, error) {
	return json.Marshal((&struct {
		Id   string `json:"id"`
		Type string `json:"type"`
	}{Id: c.client.Id, Type: "peer"}))
}

func NewPeerConversation(client *Client) PeerConversation {
	return PeerConversation{
		client: client,
	}
}

type GroupConversation struct {
	clients concurrentMap[string, *Client]
	id      string
}

func (c *GroupConversation) Id() string {
	return c.id
}

func (c *GroupConversation) AddClient(client *Client) {
	c.clients.Store(client.Id, client)
}

func (c *GroupConversation) RemoveClient(client *Client) {
	c.clients.Delete(client.Id)
}

func (c *GroupConversation) DeliverMessage(msg TextMessage) {
	c.clients.RRange(func(_ string, client *Client) bool {
		if client.Id == msg.SenderId {
			return true
		}
		newMessage := TextMessage{
			SenderId:   msg.ReceiverId,
			ReceiverId: client.Id,
			Message:    msg.Message,
			Type:       TEXT_ACTION,
		}
		client.SendTextMessage(newMessage)
		return false
	})
}

func (c *GroupConversation) MarshalJSON() ([]byte, error) {
	return json.Marshal((&struct {
		Id   string `json:"id"`
		Type string `json:"type"`
	}{Id: c.id, Type: "group"}))
}

func NewGroupConversation(clients ...*Client) GroupConversation {
	clientMap := make(map[*Client]bool)
	for _, client := range clients {
		clientMap[client] = true
	}
	return GroupConversation{
		clients: NewConcurrentMap[string, *Client](),
		id:      uuid.NewString(),
	}
}
