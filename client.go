package main

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client interface {
	SendTextMessage(TextMessage)
	Id() string
}

type LocalClient struct {
	id              string
	conn            *websocket.Conn
	messageReceiver chan any
}

func NewClient(conn *websocket.Conn) LocalClient {
	id := uuid.NewString()
	log.Println("NewClient", id)
	return LocalClient{
		id:              id,
		conn:            conn,
		messageReceiver: make(chan any),
	}
}

func (c *LocalClient) StartRead() {
	for {
		var req ClientRequestMessage
		if err := c.conn.ReadJSON(&req); err != nil {
			log.Print("Read from client error: ", err)
			gRemoveClient(c.id)
			return
		}
		go c.processRequest(req)
	}
}

func (c *LocalClient) StartWrite() {
	for message := range c.messageReceiver {
		switch msg := message.(type) {
		case TextMessage:
			if err := c.conn.WriteJSON(msg); err != nil {
				log.Println("Write TextMessage to client error: ", err)
				gRemoveClient(c.id)
			}
		case ConvListMessage:
			convs := gConvs.Values()
			err := c.conn.WriteJSON(ConvListMessage{Conversations: convs, Type: GET_CONV_LIST_ACTION})
			if err != nil {
				log.Println("Write ConvListMessage to client error: ", err)
				gRemoveClient(c.id)
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
			applied := gConvs.ApplyToOne(
				func(_ string, conv Conv) bool { return conv.Id() == msg.GroupId },
				func(_ string, conv Conv) { conv.(*GroupConv).AddClient(c) },
			)
			if !applied {
				GetRedisClient().SendMessage(NewServerRequestMessage(
					GROUP_CLIENT_JOINED,
					GroupClientJoinedMessage{
						ClientId: c.id,
						GroupId:  msg.GroupId,
						ServerId: gServerId,
					},
				))
			}
		case LeaveGroupMessage:
			applied := gConvs.ApplyToOne(
				func(_ string, conv Conv) bool { return conv.Id() == msg.GroupId },
				func(_ string, conv Conv) { conv.(*GroupConv).RemoveClient(c.id) },
			)
			if !applied {
				GetRedisClient().SendMessage(NewServerRequestMessage(
					GROUP_CLIENT_LEAVED,
					GroupClientJoinedMessage{
						ClientId: c.id,
						GroupId:  msg.GroupId,
						ServerId: gServerId,
					},
				))
			}
		}
	}
}

func (c *LocalClient) Id() string {
	return c.id
}

func (c *LocalClient) SendTextMessage(msg TextMessage) {
	go func() {
		c.messageReceiver <- msg
	}()
}

func (c *LocalClient) processRequest(req ClientRequestMessage) {
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
		c.messageReceiver <- CreateGroupMessage{Clients: []*LocalClient{c}}
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

// Remote client is used inside GroupConv
type RemoteClient struct {
	id       string
	serverId string
}

func NewRemoteClient(id, serverId string) RemoteClient {
	return RemoteClient{
		id:       id,
		serverId: serverId,
	}
}

func (c RemoteClient) Id() string {
	return c.id
}

func (c RemoteClient) SendTextMessage(msg TextMessage) {

}
