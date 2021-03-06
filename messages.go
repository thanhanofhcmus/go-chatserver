package main

const (
	// server - server actions
	TEXT_OTHER_SERVER_ACTION   = "text-other-server"
	CLIENT_CONNECTED_ACTION    = "client-connected"
	CLIENT_DISCONNECTED_ACTION = "client-disconnected"
	GROUP_CREATED_ACTION       = "group-created"
	GROUP_CLIENT_JOINED        = "group-client-joined"
	GROUP_CLIENT_LEAVED        = "group-client-leaved"

	// client - server actions
	ID_ACTION            = "id"
	TEXT_ACTION          = "text"
	GET_CONV_LIST_ACTION = "get-conversation-list"
	CREATE_GROUP_ACTION  = "create-group"
	JOIN_GROUP_ACTION    = "join-group"
	LEAVE_GROUP_ACTION   = "leave-group"
)

type ServerRequestMessage struct {
	Request        string `json:"request"`
	SenderServerId string `json:"senderServerId"`
	Data           any    `json:"data"`
}

func NewServerRequestMessage(request string, data any) ServerRequestMessage {
	return ServerRequestMessage{
		Request:        request,
		SenderServerId: gServerId,
		Data:           data,
	}
}

type ClientConnectedMessage struct {
	Id       string `json:"id"`
	ServerId string `json:"serverId"`
}

type GroupCreatedMessage struct {
	Id       string `json:"id"`
	ServerId string `json:"serverId"`
}

type GroupClientJoinedMessage struct {
	ClientId string `json:"clientId"`
	GroupId  string `json:"groupId"`
	ServerId string `json:"serverId"`
}

type GroupClientLeavedMessage struct {
	ClientId string `json:"clientId"`
	GroupId  string `json:"groupId"`
	ServerId string `json:"serverId"`
}

type ClientRequestMessage struct {
	Request string `json:"request"`
	Data    any    `json:"data"`
}

type IdMessage struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type TextMessage struct {
	SenderId   string `json:"senderId"`
	ReceiverId string `json:"receiverId"`
	Message    string `json:"message"`
	Type       string `json:"type"`
}

type ConvListMessage struct {
	Conversations []Conv `json:"conversations"`
	Type          string `json:"type"`
}

type CreateGroupMessage struct {
	Clients []*LocalClient
}

type JoinGroupMessage struct {
	SenderId string `json:"senderId"`
	GroupId  string `json:"groupId"`
}

type LeaveGroupMessage struct {
	SenderId string `json:"senderId"`
	GroupId  string `json:"groupId"`
}
