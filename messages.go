package main

type RequestMessage struct {
	Request string      `json:"request"`
	Data    interface{} `json:"data"`
}

func (r RequestMessage) String() string {
	return "{ " + r.Request + " }"
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

type ConversationListMessage struct {
	Conversations []Conversation `json:"conversations"`
	Type          string         `json:"type"`
}

type CreateGroupMessage struct {
	Clients []*Client
}

type JoinGroupMessage struct {
	SenderId string `json:"senderId"`
	GroupId  string `json:"groupId"`
}

type LeaveGroupMessage struct {
	SenderId string `json:"senderId"`
	GroupId  string `json:"groupId"`
}
