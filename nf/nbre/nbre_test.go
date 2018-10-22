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

import (
	"sync"
	"testing"

	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/stretchr/testify/assert"
)

type mockNeb struct {
	config *nebletpb.Config
}

func (n *mockNeb) Config() *nebletpb.Config {
	return n.config
}

func TestNbre_Start(t *testing.T) {
	nbre := NewNbre(nil)
	err := nbre.Start()
	assert.NoError(t, err, "nbre start failed")
	err = nbre.Shutdown()
	assert.NoError(t, err, "nbre shutdown failed")
}

func TestNbre_Execute(t *testing.T) {
	tests := []struct {
		name    string
		command string
		params  []byte
		result  []byte
		err     error
	}{
		{
			name:    "command not found",
			command: "commandNotFound",
			params:  nil,
			result:  nil,
			err:     ErrCommandNotFound,
		},
		{
			name:    "command version",
			command: CommandVersion,
			params:  nil,
			result:  nil,
			err:     nil,
		},
	}

	wg := new(sync.WaitGroup)
	wg.Add(2)

	nbre := NewNbre(nil)
	err := nbre.Start()
	assert.NoError(t, err, "nbre start failed")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer wg.Done()

			result, err := nbre.Execute(tt.command, tt.params)
			assert.Equal(t, tt.result, result)
			assert.Equal(t, tt.err, err)
		})
	}

	wg.Wait()

	err = nbre.Shutdown()
	assert.NoError(t, err, "nbre shutdown failed")
}
