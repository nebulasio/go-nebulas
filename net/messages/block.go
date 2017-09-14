package messages

import (
	"fmt"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/net"
)

const (
	NewBlockMessageType = "NewBlock"
)

type BlockMessage struct {
	t     net.MessageType
	block *core.Block
}

func NewBlockMessage(block *core.Block) *BlockMessage {
	msg := &BlockMessage{t: NewBlockMessageType, block: block}
	return msg
}

func (msg *BlockMessage) MessageType() net.MessageType {
	return msg.t
}

func (msg *BlockMessage) String() string {
	return fmt.Sprintf("BlockMessage {type:%s; block:%s}",
		msg.t,
		msg.block,
	)
}
