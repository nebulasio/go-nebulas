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

package dag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDag_AddNode(t *testing.T) {
	dag := NewDag()

	dag.AddNode("1", nil)
	dag.AddNode("2", nil)
	dag.AddNode("3", nil)
	dag.AddNode("4", nil)

	// Add the edges (Note that given Nodes must exist before adding an
	// edge between them)
	dag.AddEdge("1", "2")
	dag.AddEdge("1", "3")
	dag.AddEdge("2", "4")
	dag.AddEdge("3", "4")

	node := dag.GetNode("1")

	assert.Equal(t, "1", node.Key)
	assert.Equal(t, 0, node.ParentCounter)

	node4 := dag.GetNode("4")
	assert.Equal(t, 2, node4.ParentCounter)

	dag.AddNode("5", nil)
	msg, err := dag.ToProto()

	assert.Nil(t, err)

	err = dag.FromProto(msg)

	assert.Nil(t, err)

	node1 := dag.GetNode("5")

	assert.Equal(t, "5", node1.Key)
}
