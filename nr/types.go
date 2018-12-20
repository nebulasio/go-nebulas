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

package nr

import (
	"encoding/json"
	"errors"
)

// Error types
var (
	ErrInvalidStartHeight = errors.New("invalid nr start height")
	ErrInvalidEndHeight = errors.New("invalid nr end height")
	ErrInvalidHeightInterval   = errors.New("invalid nr height interval")
)

// const
const (
	DefaultNRVersion = 4294967296
)

// NRItem nr item
type NRItem struct {
	Address string		`json:"address"`
	Score string		`json:"score"`
}

// NRData nr data
type NRData struct {
	StartHeight uint64		`json:"start_height,string"`
	EndHeight uint64		`json:"end_height,string"`
	Version string		`json:"version"`
	Nrs []*NRItem		`json:"nrs"`
	Err string			`json:"err"`
}

// ToBytes serialize data
func (n *NRData) ToBytes() ([]byte, error) {
	return json.Marshal(n)
}

// FromBytes
func (n *NRData) FromBytes(data []byte) error {
	if err := json.Unmarshal(data, n); err != nil {
		return err
	}
	return nil
}
