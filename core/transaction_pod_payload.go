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
	"fmt"

	"github.com/nebulasio/go-nebulas/util"
)

// PoD Consensus contract functions
const (
	PoDHeartbeat = "heartbeat"
	PoDState     = "state"

	PoDMiners       = "getMiners"
	PoDCandidates   = "getCandidates"
	PoDParticipants = "getParticipants"
)

// PodPayload carry pod data
type PodPayload struct {
	Serial int64
	Action string
	Data   []byte
}

// LoadPodPayload from bytes
func LoadPodPayload(bytes []byte) (*PodPayload, error) {
	payload := &PodPayload{}
	if err := json.Unmarshal(bytes, payload); err != nil {
		return nil, ErrInvalidArgument
	}
	return NewPodPayload(payload.Serial, payload.Action, payload.Data)
}

// NewPodPayload with data
func NewPodPayload(serial int64, action string, data []byte) (*PodPayload, error) {
	if serial == 0 {
		return nil, ErrInvalidArgument
	}
	return &PodPayload{
		Serial: serial,
		Action: action,
		Data:   data,
	}, nil
}

// ToBytes serialize payload
func (payload *PodPayload) ToBytes() ([]byte, error) {
	return json.Marshal(payload)
}

// BaseGasCount returns base gas count
func (payload *PodPayload) BaseGasCount() *util.Uint128 {
	return util.NewUint128()
}

func (payload *PodPayload) heartbeat() (string, string, error) {
	args := fmt.Sprintf("[%d]", payload.Serial)
	return payload.Action, args, nil
}

func (payload *PodPayload) state(block *Block) (string, string, error) {
	args := fmt.Sprintf("[%d]", payload.Serial)
	return payload.Action, args, nil
}

// Execute the payload in tx
func (payload *PodPayload) Execute(limitedGas *util.Uint128, tx *Transaction, block *Block, ws WorldState) (*util.Uint128, string, error) {
	if block == nil || tx == nil {
		return util.NewUint128(), "", ErrNilArgument
	}

	var (
		function string
		args     string
		err      error
	)
	switch payload.Action {
	case PoDHeartbeat:
		function, args, err = payload.heartbeat()
	case PoDState:
		function, args, err = payload.state(block)
	default:
		err = ErrInvalidArgument
	}
	if err != nil {
		return util.NewUint128(), "", err
	}

	instructions, result, exeErr := executeContract(limitedGas, tx, block, ws, function, args)
	if exeErr == ErrExecutionFailed && len(result) > 0 {
		exeErr = fmt.Errorf("PoD: %s", result)
	}
	if exeErr != nil {
		return instructions, "", exeErr
	}

	return instructions, result, exeErr
}
