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
	"errors"
	"encoding/json"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/util"
)

// Error types
var (
	ErrInvalidHeight = errors.New("invalid dip height")
	ErrDipNotFound = errors.New("dip not found")

	ErrInvalidDipAddress                    = errors.New("invalid dip reward address")
	ErrUnsupportedTransactionFromDipAddress = errors.New("unsupported transaction from dip address")

)

// const types
const (
	CacheSize = 16
	DipDelayRewardHeight = 24*60*60/15

	// DipRewardAddressPrivate dip reward rewardAddress
	DipRewardAddressPrivate = "42f0c8b5feb72301619046ca87e6cf2c605e94dae0e24c9cb3a0101dbb60337c"
	DipRewardAddressPassphrase = "passphrase"

)

var (
	// BlockReward given to dip address
	// rule: 0.4% of block reward
	// value: 1.42694 * 10^18/1000*4 = 5.70776e+15
	DipRewardValue, _ = core.BlockReward.Div(util.NewUint128FromUint(250))
)

type DIPItem struct {
	Address  string		`json:"rewardAddress"`
	Reward string		`json:"reward"`
}

type DIPData struct {
	Start uint64		`json:"start"`
	End uint64			`json:"end"`
	Version string		`json:"version"`
	Dips []*DIPItem		`json:"dips"`
	Err string			`json:"err"`
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