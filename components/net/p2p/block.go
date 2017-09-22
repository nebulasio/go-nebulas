// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package p2p

import (
	"time"

	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	nnet "github.com/nebulasio/go-nebulas/components/net"
	"github.com/nebulasio/go-nebulas/components/net/messages"
	"github.com/nebulasio/go-nebulas/core"
	b "github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

const blockProtocolID = "/nebulas/block/1.0.0"

type BlockMsgService struct {
	node *Node
	np   *P2pManager
}

func (np *P2pManager) RegisterBlockMsgService() *BlockMsgService {
	bs := &BlockMsgService{np.node, np}
	np.node.host.SetStreamHandler(blockProtocolID, np.BlockMsgHandler)
	log.Infof("RegisterBlockMsgService: node register block message service success...")
	return bs
}

func (np *P2pManager) BlockMsgHandler(s net.Stream) {
	defer s.Close()
	log.Info("BlockMsgHandler: handle block msg ")
	timeout := 30 * time.Second
	size, err := ReadWithTimeout(s, 4, timeout)
	data, err := ReadWithTimeout(
		s,
		b.Uint32(size),
		timeout,
	)

	block := new(core.Block)
	err = block.Deserialize(data)
	if err != nil {
		log.Error("BlockMsgHandler: handle block msg occurs error: ", err)
	}
	msg := messages.NewBaseMessage(nnet.MessageTypeNewBlock, block)
	log.Info("BlockMsgHandler: handle block msg -> ", msg)
	np.PutMessage(msg)
}

func (node *Node) SendBlock(msg *core.Block, pid peer.ID) {

	log.Info("SendBlock: send block msg to", pid, msg)
	stream, err := node.host.NewStream(node.context, pid, blockProtocolID)
	if err != nil {
		log.Error("SendBlock: send block msg occurs error ", err)
		return
	}
	defer stream.Close()

	//var s b.Serializable = &b.JSONSerializer{}
	//data, err := s.Serialize(*msg)
	data, err := msg.Serialize()

	if err != nil {
		return
	}
	timeout := 30 * time.Second
	size := b.FromUint32(uint32(len(data)))
	err = WriteWithTimeout(
		stream,
		append(size[:], data...),
		timeout,
	)
}
