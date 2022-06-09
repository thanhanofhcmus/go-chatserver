package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	TEXT_ACTION                  = "text"
	GET_CONVERSATION_LIST_ACTION = "get-conversation-list"
	CREATE_GROUP_ACTION          = "create-group"
	JOIN_GROUP_ACTION            = "join-group"
	LEAVE_GROUP_ACTION           = "leave-group"
)

type Client struct {
	Id              string          `json:"id"`
	Conn            *websocket.Conn `json:"-"`
	messageReceiver chan interface{}
}

func NewClient(conn *websocket.Conn) Client {
	id := uuid.NewString()
	log.Println("NewClient", id)
	return Client{
		Id:              id,
		Conn:            conn,
		messageReceiver: make(chan interface{}),
	}
}

func (c *Client) String() string {
	return fmt.Sprintf("U{ %s }", c.Id)
}

func (c *Client) StartRead() {
	for {
		var req RequestMessage
		if err := c.Conn.ReadJSON(&req); err != nil {
			log.Print("read error: " + err.Error())
			gClientRemover <- c
			return
		}
		c.processRequest(req)
	}
}

func (c *Client) StartWrite() {
	for {
		for message := range c.messageReceiver {
			switch msg := message.(type) {
			case TextMessage:
				if err := c.Conn.WriteJSON(msg); err != nil {
					log.Println(err)
					gClientRemover <- c
				}
			case ConversationListMessage:
				keys := make([]Conversation, 0, gConversations.Count())
				gConversations.RRange(func(key string, c Conversation) bool {
					keys = append(keys, c)
					return true
				})
				err := c.Conn.WriteJSON(ConversationListMessage{Conversations: keys, Type: "get-conversation-list"})
				if err != nil {
					log.Println(err)
					gClientRemover <- c
				}
			case CreateGroupMessage:
				conv := NewGroupConversation(msg.Clients...)
				gConversations.Store(conv.Id(), conv)
			case JoinGroupMessage:
				gConversations.Range(func(_ string, conv Conversation) bool {
					if conv.Id() == msg.GroupId {
						conv := conv.(GroupConversation)
						conv.AddClient(c)
						return false
					}
					return true
				})
			case LeaveGroupMessage:
				gConversations.Range(func(_ string, conv Conversation) bool {
					if conv.Id() == msg.GroupId {
						conv := conv.(GroupConversation)
						conv.RemoveClient(c)
						return false
					}
					return true
				})
			}
		}
	}
}

func (c *Client) SendTextMessage(msg TextMessage) {
	c.messageReceiver <- msg
}

func (c *Client) processRequest(req RequestMessage) {
	log.Println(req)
	switch req.Request {
	case TEXT_ACTION:
		bytes, err := json.Marshal(req.Data)
		if err != nil {
			log.Print("parse JSON from req.Data in text message error: " + err.Error())
			return
		}
		var msg TextMessage
		if err := json.Unmarshal(bytes, &msg); err != nil {
			log.Print("parse JSON to TexMessage in text message error: " + err.Error())
			return
		}
		gConversations.RRange(func(_ string, conv Conversation) bool {
			if conv.Id() == msg.ReceiverId {
				conv.DeliverMessage(msg)
				return false
			}
			return true
		})
	case GET_CONVERSATION_LIST_ACTION:
		c.messageReceiver <- ConversationListMessage{}
	case CREATE_GROUP_ACTION:
		c.messageReceiver <- CreateGroupMessage{Clients: []*Client{c}}
	case JOIN_GROUP_ACTION:
		bytes, err := json.Marshal(req.Data)
		if err != nil {
			log.Print("parse JSON from req.Data in join group message error: " + err.Error())
			return
		}
		var msg JoinGroupMessage
		if err := json.Unmarshal(bytes, &msg); err != nil {
			log.Print("parse JSON to JoinGroupMessage in text message error: " + err.Error())
			return
		}
		c.messageReceiver <- msg
	case LEAVE_GROUP_ACTION:
		bytes, err := json.Marshal(req.Data)
		if err != nil {
			log.Print("parse JSON from req.Data in join group message error: " + err.Error())
			return
		}
		var msg LeaveGroupMessage
		if err := json.Unmarshal(bytes, &msg); err != nil {
			log.Print("parse JSON to JoinGroupMessage in text message error: " + err.Error())
			return
		}
		c.messageReceiver <- msg
	}
}
