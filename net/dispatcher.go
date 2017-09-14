package net

import (
	log "github.com/sirupsen/logrus"
)

type Dispatcher struct {
	messageHandlerMap MessageHandlerMap

	quitCh            chan bool
	receivedMessageCh chan Message
}

type MessageHandlerMap map[MessageType]MessageHandlers

func NewDispatcher() *Dispatcher {
	dp := &Dispatcher{
		messageHandlerMap: make(MessageHandlerMap),

		quitCh:            make(chan bool, 10),
		receivedMessageCh: make(chan Message, 1024),
	}

	return dp
}

func (dp *Dispatcher) Register(handler MessageHandler) {
	for _, v := range handler.SubscribeMessageTypes() {
		msgHandlers := dp.messageHandlerMap[v]
		if msgHandlers == nil {
			msgHandlers = make(MessageHandlers)
			dp.messageHandlerMap[v] = msgHandlers
		}
		msgHandlers[handler] = true
	}
}

func (dp *Dispatcher) Deregister(handler MessageHandler) {
	for _, v := range handler.SubscribeMessageTypes() {
		msgHandlers := dp.messageHandlerMap[v]
		if msgHandlers == nil {
			continue
		}
		delete(msgHandlers, handler)
		if len(msgHandlers) == 0 {
			delete(dp.messageHandlerMap, v)
		}
	}
}

func (dp *Dispatcher) Start() {
	go (func() {
		for {
			select {
			case <-dp.quitCh:
				log.Info("dispatcher is stopped.")
				return
			case msg := <-dp.receivedMessageCh:
				msgID := msg.MessageType()
				msgHandlers := dp.messageHandlerMap[msgID]
				if msgHandlers == nil || len(msgHandlers) == 0 {
					log.WithFields(log.Fields{
						"MessageType": msgID,
					}).Info("receive message without handler.")
					break
				}

				for handler := range msgHandlers {
					handler.OnMessageReceived(msg)
				}
			}
		}
	})()
}

func (dp *Dispatcher) Stop() {
	dp.quitCh <- true
}

func (dp *Dispatcher) OnMessageReceived(msg Message) {
	dp.receivedMessageCh <- msg
}
