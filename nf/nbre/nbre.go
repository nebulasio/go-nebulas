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

package nbre

import "github.com/nebulasio/go-nebulas/core"

// Nbre type of Nbre
type Nbre struct{}

// NewNbre create new Nbre
func NewNbre() core.Nbre {
	return &Nbre{}
}

// Start launch the nbre
func (s *Nbre) Start() error {
	//TODO(larry): start the nbre
	return nil
}

// Execute execute command
func (s *Nbre) Execute(command string, params []byte) ([]byte, error) {
	//TODO(larry): add execute for nbre
	return nil, nil
}
