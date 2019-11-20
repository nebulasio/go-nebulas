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

	"github.com/gogo/protobuf/proto"
	dagpb "github.com/nebulasio/go-nebulas/common/dag/pb"
	"github.com/stretchr/testify/assert"
)

func TestDagProto(t *testing.T) {
	d := NewDag()
	var pd *dagpb.Dag
	assert.Equal(t, d.FromProto(pd), ErrInvalidProtoToDag)
}

func TestDag_AddNode(t *testing.T) {
	dag := NewDag()

	dag.AddNode("1")
	dag.AddNode("2")
	dag.AddNode("3")
	dag.AddNode("4")

	// Add the edges (Note that given Nodes must exist before adding an
	// edge between them)
	dag.AddEdge("1", "2")
	dag.AddEdge("1", "3")
	dag.AddEdge("2", "4")
	dag.AddEdge("3", "4")

	node := dag.GetNode("1")

	assert.Equal(t, "1", node.key)
	assert.Equal(t, 0, node.index)

	assert.Equal(t, 0, node.parentCounter)

	node4 := dag.GetNode("4")
	assert.Equal(t, 2, node4.parentCounter)

	dag.AddNode("5")
	msg, err := dag.ToProto()

	assert.Nil(t, err)

	node1 := dag.GetNode("5")

	assert.Equal(t, "5", node1.key)

	msg, err1 := dag.ToProto()
	assert.Nil(t, err1)

	dag1 := NewDag()
	err = dag1.FromProto(msg)
	assert.Nil(t, err)
	node1 = dag1.GetNode(int(0))

	//fmt.Println(err, node1)
	assert.Equal(t, 0, node1.key)

	children := dag1.GetChildrenNodes(int(0))

	assert.Equal(t, 2, len(children))
}

func TestDag_ToFromProto(t *testing.T) {
	dag1 := NewDag()

	dag1.AddNode("key1")
	dag1.AddNode("key2")
	dag1.AddNode("key3")
	dag1.AddNode("key4")
	dag1.AddNode("key5")

	dag1.AddEdge("key1", "key2")
	dag1.AddEdge("key1", "key3")
	dag1.AddEdge("key2", "key4")
	dag1.AddEdge("key3", "key4")

	msg1, err := dag1.ToProto()
	assert.Nil(t, err)

	j1, err := proto.Marshal(msg1)
	assert.Nil(t, err)

	dag2 := NewDag()
	err = dag2.FromProto(msg1)
	assert.Nil(t, err)

	msg2, err := dag2.ToProto()
	assert.Nil(t, err)
	j2, err := proto.Marshal(msg2)
	assert.Nil(t, err)

	assert.Equal(t, j1, j2)

}
func TestDag_IsCirclular(t *testing.T) {
	dag := NewDag()

	dag.AddNode("1")
	dag.AddNode("2")
	dag.AddNode("3")
	dag.AddNode("4")
	dag.AddNode("5")

	// Add the edges (Note that given Nodes must exist before adding an
	// edge between them)
	dag.AddEdge("1", "2")
	dag.AddEdge("1", "3")
	dag.AddEdge("2", "4")
	dag.AddEdge("3", "4")

	assert.Equal(t, false, dag.IsCirclular())
	_, err := dag.ToProto()
	assert.Nil(t, err)

	dag.AddEdge("4", "1")

	assert.Equal(t, true, dag.IsCirclular())

	msg, err := dag.ToProto()
	assert.Nil(t, err)

	dag1 := NewDag()
	err = dag1.FromProto(msg)
	assert.Nil(t, err)

	dag.AddEdge("4", "5")
	dag.AddEdge("5", "1")
	assert.Equal(t, true, dag.IsCirclular())

}

func TestDag_IsCirclular1(t *testing.T) {

	dag := NewDag()

	dag.AddNode("1")
	dag.AddNode("2")
	dag.AddNode("3")
	dag.AddNode("4")
	dag.AddNode("5")
	dag.AddNode("6")
	dag.AddNode("7")
	dag.AddNode("8")
	dag.AddNode("9")
	dag.AddNode("10")
	dag.AddNode("11")
	dag.AddNode("12")
	dag.AddNode("13")
	dag.AddNode("14")
	dag.AddNode("15")
	dag.AddNode("16")
	dag.AddNode("17")
	dag.AddNode("18")
	dag.AddNode("19")
	// Add the edges (Note that given vertices must exist before adding an
	// edge between them)
	dag.AddEdge("1", "2")
	dag.AddEdge("1", "3")
	dag.AddEdge("2", "4")
	dag.AddEdge("3", "5")
	dag.AddEdge("3", "6")
	dag.AddEdge("4", "7")
	dag.AddEdge("5", "8")
	dag.AddEdge("5", "9")
	dag.AddEdge("6", "10")
	dag.AddEdge("6", "11")
	dag.AddEdge("9", "12")
	dag.AddEdge("10", "12")
	dag.AddEdge("12", "13")
	dag.AddEdge("13", "15")
	dag.AddEdge("12", "14")

	dag.AddEdge("16", "17")
	dag.AddEdge("16", "18")
	dag.AddEdge("18", "19")

	assert.Equal(t, false, dag.IsCirclular())

	dag.AddEdge("15", "8")
	assert.Equal(t, false, dag.IsCirclular())

	dag.AddEdge("19", "16")
	assert.Equal(t, true, dag.IsCirclular())
}
