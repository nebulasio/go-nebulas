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
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// PoD Consensus contract functions
const (
	PoDHeartbeat = "heartbeat"
	PoDState     = "state"
	PoDReport    = "witness"

	PoDMiners       = "getMiners"
	PoDCandidates   = "getCandidates"
	PoDParticipants = "getParticipants"
	PoDNodeInfo     = "getNodeInfo"
)

const (
	AttackDoubleSpend = "doubleSpend"
	AttackNotMiner    = "notMiner"
)

type Report struct {
	Timestamp int64  `json:"timestamp"`
	Miner     string `json:"miner"`
	Evil      string `json:"evil"`
}

func (w *Report) ToBytes() ([]byte, error) {
	return json.Marshal(w)
}

func (w *Report) FromBytes(data []byte) error {
	if err := json.Unmarshal(data, w); err != nil {
		return err
	}
	return nil
}

type Statistics struct {
	Serial     int64          `json:"serial"`
	Start      uint64         `json:"start"`
	Statistics map[string]int `json:"statistics"`
}

func (s *Statistics) Equals(state *Statistics) bool {
	if s.Serial == state.Serial {
		if s.Start != state.Start {
			return false
		}
		if s.Statistics == nil && state.Statistics == nil {
			return true
		}
		if len(s.Statistics) != len(state.Statistics) {
			return false
		}
		for k, v := range s.Statistics {
			if state.Statistics[k] != v {
				return false
			}
		}
		return true
	}
	return false
}

func (s *Statistics) ToBytes() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Statistics) FromBytes(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	return nil
}

type NodeInfo struct {
	Id              string `json:"id"`
	HeartbeatSerial int64  `json:"heartbeat_serial"`
	Miner           string `json:"miner"`
	Score           string `json:"score"`
}

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

// heartbeat submit participants heartbeat
func (payload *PodPayload) heartbeat() (string, string, error) {
	args := fmt.Sprintf("[%d]", payload.Serial)
	return payload.Action, args, nil
}

// state submit pod state
func (payload *PodPayload) state(tx *Transaction, block *Block) (string, string, error) {
	var (
		states []*Statistics
		err    error
	)
	if err = json.Unmarshal(payload.Data, &states); err != nil {
		return "", "", err
	}

	//for _, v := range states {
	//	logging.VLog().WithFields(logrus.Fields{
	//		"tx.hash": tx.Hash(),
	//		"states":  v,
	//	}).Debug("load pod statistics")
	//}

	blockStates, err := block.txPool.bc.StatisticalLastBlocks(payload.Serial, block)
	if err != nil {
		return "", "", err
	}
	if len(states) != len(blockStates) {
		return "", "", ErrBlockStateCheckFailed
	}
	for _, blockState := range blockStates {
		found := false
		for _, state := range states {
			if blockState.Equals(state) {
				found = true
				break
			}
		}
		if !found {
			err = ErrBlockStateCheckFailed
			logging.VLog().WithFields(logrus.Fields{
				"tx.hash":    tx.Hash(),
				"blockState": blockState,
				"count":      len(blockStates),
			}).Error("Failed to check block state statistics")
			break
		}
	}

	if err != nil {
		return "", "", err
	}

	// args serialize
	args := make([]interface{}, 2)
	args[0] = payload.Serial + 1
	args[1] = states
	bytes, err := json.Marshal(args)
	if err != nil {
		return "", "", err
	}
	return payload.Action, string(bytes), nil
}

// report submit node evil
func (payload *PodPayload) report(block *Block) (string, string, error) {
	report := new(Report)
	if err := report.FromBytes(payload.Data); err != nil {
		return "", "", nil
	}
	args := fmt.Sprintf("[%d, %s, %s]", report.Timestamp, report.Miner, report.Evil)
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
		function, args, err = payload.state(tx, block)
	case PoDReport:
		function, args, err = payload.report(block)
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
