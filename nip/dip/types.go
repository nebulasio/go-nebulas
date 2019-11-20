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

package dip

import (
	"encoding/json"
	"errors"

	"github.com/nebulasio/go-nebulas/core"
	nebletpb "github.com/nebulasio/go-nebulas/neblet/pb"
)

// Error types
var (
	ErrDipNotFound = errors.New("dip not found")

	ErrInvalidDipAddress                    = errors.New("invalid dip reward address")
	ErrUnsupportedTransactionFromDipAddress = errors.New("unsupported transaction from dip address")
)

// const types
const (
	CacheSize = 128

	// DipRewardAddressPrivate dip reward rewardAddress:n1c6y4ctkMeZk624QWBTXuywmNpCWmJZiBq
	DipRewardAddressPrivate = "42f0c8b5feb72301619046ca87e6cf2c605e94dae0e24c9cb3a0101dbb60337c"
	// DipRewardAddressPassphrase
	DipRewardAddressPassphrase = "passphrase"
)

type Neblet interface {
	Config() *nebletpb.Config
	AccountManager() core.AccountManager
	BlockChain() *core.BlockChain
}

type DipReward struct {
	DipRewards []string `json:"dip_rewards"`
}

type DIPItem struct {
	Address  string `json:"address"`
	Contract string `json:"contract"`
	Reward   string `json:"reward"`
}

type DIPData struct {
	StartHeight uint64     `json:"start_height,string"`
	EndHeight   uint64     `json:"end_height,string"`
	Version     uint64     `json:"version,string"`
	Dips        []*DIPItem `json:"dips"`
	Err         string     `json:"err"`
}

// ToBytes serialize data
func (d *DIPData) ToBytes() ([]byte, error) {
	return json.Marshal(d)
}

// FromBytes
func (d *DIPData) FromBytes(data []byte) error {
	if err := json.Unmarshal(data, d); err != nil {
		return err
	}
	return nil
}
