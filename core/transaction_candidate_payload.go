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

	"github.com/nebulasio/go-nebulas/util"
	// "github.com/nebulasio/go-nebulas/util/logging"
	// "github.com/sirupsen/logrus"
)

// Candidate Action
const (
	LoginAction  = "login"
	LogoutAction = "logout"
)

// CandidatePayload carry candidate application
type CandidatePayload struct {
	Action string
}

// LoadCandidatePayload from bytes
func LoadCandidatePayload(bytes []byte) (*CandidatePayload, error) {
	payload := &CandidatePayload{}
	if err := json.Unmarshal(bytes, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// NewCandidatePayload with comments
func NewCandidatePayload(action string) *CandidatePayload {
	return &CandidatePayload{
		Action: action,
	}
}

// ToBytes serialize payload
func (payload *CandidatePayload) ToBytes() ([]byte, error) {
	return json.Marshal(payload)
}

// BaseGasCount returns base gas count
func (payload *CandidatePayload) BaseGasCount() *util.Uint128 {
	return CandidateBaseGasCount
}

// Execute the candidate payload in tx
func (payload *CandidatePayload) Execute(block *Block, tx *Transaction) (*util.Uint128, string, error) {
	candidate := tx.from.Bytes()
	switch payload.Action {
	case LoginAction:
		if _, err := block.dposContext.candidateTrie.Put(candidate, candidate); err != nil {
			return ZeroGasCount, "", err
		}
		/* 		logging.VLog().WithFields(logrus.Fields{
			"block":     ctx.block,
			"tx":        ctx.tx,
			"candidate": ctx.tx.from.String(),
		}).Debug("Candidate login.") */
	case LogoutAction:
		if err := block.dposContext.kickoutCandidate(candidate); err != nil {
			return ZeroGasCount, "", err
		}
		/* 		logging.VLog().WithFields(logrus.Fields{
			"block":     ctx.block,
			"tx":        ctx.tx,
			"candidate": ctx.tx.from.String(),
		}).Debug("Candidate logout.") */
	default:
		return ZeroGasCount, "", ErrInvalidCandidatePayloadAction
	}
	return ZeroGasCount, "", nil
}
