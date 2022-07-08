package main

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	Id              string          `json:"id"`
	conn            *websocket.Conn `json:"-"`
	messageReceiver chan any
}

func NewClient(conn *websocket.Conn) Client {
	id := uuid.NewString()
	log.Println("NewClient", id)
	return Client{
		Id:              id,
		conn:            conn,
		messageReceiver: make(chan any),
	}
}

func (c *Client) StartRead() {
	for {
		var req ClientRequestMessage
		if err := c.conn.ReadJSON(&req); err != nil {
			log.Print("Read from client error: ", err)
			gRemoveClient(c.Id)
			return
		}
		go c.processRequest(req)
	}
}

func (c *Client) StartWrite() {
	for message := range c.messageReceiver {
		switch msg := message.(type) {
		case TextMessage:
			if err := c.conn.WriteJSON(msg); err != nil {
				log.Println("Write TextMessage to client error: ", err)
				gRemoveClient(c.Id)
			}
		case ConvListMessage:
			convs := gConvs.Values()
			err := c.conn.WriteJSON(ConvListMessage{Conversations: convs, Type: GET_CONV_LIST_ACTION})
			if err != nil {
				log.Println("Write ConvListMessage to client error: ", err)
				gRemoveClient(c.Id)
			}
		case CreateGroupMessage:
			conv := NewGroupConv(msg.Clients...)
			gConvs.Store(conv.Id(), conv)
			gcMsg := GroupCreatedMessage{
				Id:       conv.id,
				ServerId: gServerId,
			}
			GetRedisClient().SendMessage(NewServerRequestMessage(GROUP_CREATED_ACTION, gcMsg))
		case JoinGroupMessage:
			gConvs.ApplyToOne(
				func(_ string, conv Conv) bool { return conv.Id() == msg.GroupId },
				func(_ string, conv Conv) { conv.(*GroupConv).AddClient(c) },
			)
		case LeaveGroupMessage:
			gConvs.ApplyToOne(
				func(_ string, conv Conv) bool { return conv.Id() == msg.GroupId },
				func(_ string, conv Conv) { conv.(*GroupConv).RemoveClient(c.Id) },
			)
		}
	}
}

func (c *Client) SendTextMessage(msg TextMessage) {
	go func() {
		c.messageReceiver <- msg
	}()
}

func (c *Client) processRequest(req ClientRequestMessage) {
	log.Println(req)
	switch req.Request {
	case TEXT_ACTION:
		if msg, ok := marshalJSON[TextMessage](req.Data); ok {
			applied := gConvs.RApplyToOne(
				func(_ string, conv Conv) bool { return conv.Id() == msg.ReceiverId },
				func(_ string, conv Conv) { conv.DeliverTextMessage(msg) },
			)
			if !applied {
				GetRedisClient().SendMessage(NewServerRequestMessage(TEXT_OTHER_SERVER_ACTION, msg))
			}
		}
	case GET_CONV_LIST_ACTION:
		c.messageReceiver <- ConvListMessage{}
	case CREATE_GROUP_ACTION:
		c.messageReceiver <- CreateGroupMessage{Clients: []*Client{c}}
	case JOIN_GROUP_ACTION:
		if msg, ok := marshalJSON[JoinGroupMessage](req.Data); ok {
			c.messageReceiver <- msg
		}
	case LEAVE_GROUP_ACTION:
		if msg, ok := marshalJSON[LeaveGroupMessage](req.Data); ok {
			c.messageReceiver <- msg
		}
	}
}

func marshalJSON[T any](data any) (res T, ok bool) {
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Marshal JSON %T error: %s\n", data, err)
		return
	}
	if err := json.Unmarshal(bytes, &res); err != nil {
		log.Printf("Unmarshal JSON to %T error: %s\n", res, err)
		return
	}
	ok = true
	return
}
