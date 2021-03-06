package main

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
)

const (
	PEER_TYPE  = "peer"
	GROUP_TYPE = "group"
)

type Conv interface {
	Id() string
	ServerId() string
	DeliverTextMessage(TextMessage)
}

type PeerConv struct {
	client   *LocalClient
	serverId string
}

func (c PeerConv) Id() string {
	return c.client.id
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
	}{Id: c.client.id, ServerId: c.serverId, Type: PEER_TYPE}))
}

func NewPeerConv(client *LocalClient) PeerConv {
	return PeerConv{client: client, serverId: gServerId}
}

type GroupConv struct {
	clients       concurrentMap[string, Client]
	id            string
	serverId      string
	clientRemover chan string
}

func (c *GroupConv) Id() string {
	return c.id
}

func (c *GroupConv) ServerId() string {
	return c.serverId
}

func (c *GroupConv) AddClient(client Client) {
	c.clients.Store(client.Id(), client)
}

func (c *GroupConv) RemoveClient(clientId string) {
	go func() {
		c.clientRemover <- clientId
	}()
}

func (c *GroupConv) DeliverTextMessage(msg TextMessage) {
	c.clients.RRange(func(_ string, client Client) bool {
		if client.Id() != msg.SenderId {
			newMessage := TextMessage{
				SenderId:   msg.SenderId,
				ReceiverId: client.Id(),
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
	}{Id: c.id, ServerId: c.serverId, Type: GROUP_TYPE}))
}

func NewGroupConv(clients ...*LocalClient) *GroupConv {
	clientMap := make(map[*LocalClient]bool)
	for _, client := range clients {
		clientMap[client] = true
	}
	conv := &GroupConv{
		clients:       NewConcurrentMap[string, Client](),
		id:            uuid.NewString(),
		serverId:      gServerId,
		clientRemover: make(chan string),
	}
	// start this group remove client goroutine
	go func() {
		for clientId := range conv.clientRemover {
			conv.clients.Delete(clientId)
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
	log.Printf("%s send message %s to client %s\n", gServerId, msg.Message, c.ID)
	GetRedisClient().SendMessage(NewServerRequestMessage(TEXT_OTHER_SERVER_ACTION, msg))
}
