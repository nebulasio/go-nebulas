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

package core

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/nebulasio/go-nebulas/nf/nvm"
)

const (
	TxPayloadBinaryType = "binary"
	TxPayloadDeployType = "deploy"
	TxPayloadCallType   = "call"
	TxPayloadVoteType   = "vote"
)

var (
	ErrInvalidTxPayloadType = errors.New("invalid transaction data payload type")
	v8engineOnce            = &sync.Once{}
)

type txPayload struct {
	payloadType string `json: "type"`
	source      string `json: "source"`
	function    string `json: "function"`
	args        string `json: "args"`

	binaryData []byte
}

func (payload *txPayload) Execute(tx *Transaction, block *Block) error {
	v8engineOnce.Do(func() {
		nvm.InitV8Engine()
	})

	engine := nvm.NewV8Engine()
	defer engine.Delete()

	switch payload.payloadType {
	case TxPayloadBinaryType:
		return nil
	case TxPayloadDeployType:
		return engine.DeployAndInit(payload.source, payload.args)
	case TxPayloadCallType:
		return engine.Call(payload.function, payload.args)
	default:
		return ErrInvalidTxPayloadType
	}
}

func NewDeploySCPayload(source, args string) ([]byte, error) {
	payload := &txPayload{
		payloadType: TxPayloadDeployType,
		source:      source,
		args:        args,
	}
	return json.Marshal(payload)
}

func NewCallSCPayload(function, args string) ([]byte, error) {
	payload := &txPayload{
		payloadType: TxPayloadCallType,
		function:    function,
		args:        args,
	}
	return json.Marshal(payload)
}

func ParseTxPayload(data []byte) (*txPayload, error) {
	payload := &txPayload{}
	if err := json.Unmarshal(data, &payload); err != nil {
		payload.payloadType = TxPayloadBinaryType
		payload.binaryData = data
	}
	return payload, nil
}
