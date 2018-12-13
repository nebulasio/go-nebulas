// Copyright (C) 2018 go-nebulas authors
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

package nbre

import (
	"errors"

	"encoding/json"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/neblet/pb"
)

// Error types
var (
	ErrConfigNotFound   = errors.New("nbre config not found")
	ErrNbreStartFailed  = errors.New("nbre start failed")
	ErrCommandNotFound  = errors.New("nbre command not found")
	ErrExecutionTimeout = errors.New("nbre execute timeout")
	ErrHandlerNotFound  = errors.New("nbre handler not found")
)

// Command types
var (
	CommandVersion    = "version"
	CommandIRList     = "irList"
	CommandIRVersions = "irVersions"
	CommandNRHash     = "nrHash"
	CommandNRList     = "nrList"
	CommandDIPList    = "dipList"
)

type Neblet interface {
	Config() *nebletpb.Config
	BlockChain() *core.BlockChain
}

type Version struct {
	Major uint64 `json:"major"`
	Minor uint64 `json:"minor"`
	Patch uint64 `json:"patch"`
}

func (v *Version) ToBytes() ([]byte, error) {
	return json.Marshal(v)
}

func (v *Version) FromBytes(data []byte) error {
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	return nil
}

type Params struct {
	StartBlock uint64 `json:"start_block"`
	EndBlock   uint64 `json:"end_block"`
	Version    uint64 `json:"version"`
}

func (p *Params) ToBytes() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Params) FromBytes(data []byte) error {
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}
	return nil
}
