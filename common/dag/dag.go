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
	"errors"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/dag/pb"
)

// Node struct
type Node struct {
	Key           string
	Value         interface{}
	Children      []*Node
	ParentCounter int
}

var (
	ErrKeyNotFound  = errors.New("not found")
	ErrKeyIsExisted = errors.New("already existed")
)

// NewNode new node
func NewNode(key string, value interface{}) *Node {
	return &Node{
		Key:           key,
		Value:         value,
		ParentCounter: 0,
		Children:      make([]*Node, 0),
	}
}

// Dag struct
type Dag struct {
	Nodes map[string]*Node
}

// ToProto converts domain Dag into proto Dag
func (dag *Dag) ToProto() (proto.Message, error) {
	nodes := make([]*dagpb.Node, len(dag.Nodes))

	idx := 0
	for _, v := range dag.Nodes {
		node := new(dagpb.Node)
		node.Key = v.Key
		node.Children = make([]string, len(v.Children))
		for i, child := range v.Children {
			node.Children[i] = child.Key
		}

		nodes[idx] = node
		idx++
	}

	return &dagpb.Dag{
		Nodes: nodes,
	}, nil
}

// FromProto converts proto Dag to domain Dag
func (dag *Dag) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*dagpb.Dag); ok {
		//dag.cont
		dag.Nodes = make(map[string]*Node, len(msg.Nodes))

		for _, v := range msg.Nodes {
			dag.AddNode(v.Key, nil)
		}

		for _, v := range msg.Nodes {
			for _, child := range v.Children {
				dag.AddEdge(v.Key, child)
			}
		}

		return nil
	}
	return errors.New("Protobuf message cannot be converted into Dag")
}

// NewDag new dag
func NewDag() *Dag {
	return &Dag{
		Nodes: make(map[string]*Node, 0),
	}
}

// Len Dag len
func (dag *Dag) Len() int {
	return len(dag.Nodes)
}

// GetNode get node by key
func (dag *Dag) GetNode(key string) *Node {
	if v, ok := dag.Nodes[key]; ok {
		return v
	}
	return nil
}

// GetChildrenNodes get children nodes with key
func (dag *Dag) GetChildrenNodes(key string) []*Node {
	if v, ok := dag.Nodes[key]; ok {
		return v.Children
	}
	return nil
}

// GetRootNodes get root nodes
func (dag *Dag) GetRootNodes() []*Node {
	nodes := make([]*Node, 0)
	for _, node := range dag.Nodes {
		if node.ParentCounter == 0 {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetNodes get all nodes
func (dag *Dag) GetNodes() []*Node {
	nodes := make([]*Node, 0)
	for _, node := range dag.Nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// AddNode add node
func (dag *Dag) AddNode(key string, value interface{}) error {
	if _, ok := dag.Nodes[key]; ok {
		return ErrKeyIsExisted
	}

	dag.Nodes[key] = NewNode(key, value)
	return nil
}

// AddEdge add edge fromKey toKey
func (dag *Dag) AddEdge(fromKey, toKey string) error {
	var from, to *Node
	var ok bool

	if from, ok = dag.Nodes[fromKey]; !ok {
		return ErrKeyNotFound
	}

	if to, ok = dag.Nodes[toKey]; !ok {
		return ErrKeyNotFound
	}

	for _, childNode := range from.Children {
		if childNode == to {
			return ErrKeyIsExisted //todo
		}
	}

	dag.Nodes[toKey].ParentCounter++
	from.Children = append(from.Children, to)
	return nil
}
