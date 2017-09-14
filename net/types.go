package net

type MessageType string

type Message interface {
	MessageType() MessageType
}
type MessageHandler interface {
	SubscribeMessageTypes() []MessageType
	OnMessageReceived(msg Message)
}

type MessageHandlers map[MessageHandler]bool
