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
	"encoding/json"
	"errors"

	"github.com/gogo/protobuf/proto"
	dagpb "github.com/nebulasio/go-nebulas/common/dag/pb"
)

// Node struct
type Node struct {
	key           interface{}
	index         int
	children      []*Node
	parentCounter int
}

// Errors
var (
	ErrKeyNotFound       = errors.New("not found")
	ErrKeyIsExisted      = errors.New("already existed")
	ErrInvalidProtoToDag = errors.New("Protobuf message cannot be converted into Dag")
	ErrInvalidDagToProto = errors.New("Dag cannot be converted into Protobuf message")
)

// NewNode new node
func NewNode(key interface{}, index int) *Node {
	return &Node{
		key:           key,
		index:         index,
		parentCounter: 0,
		children:      make([]*Node, 0),
	}
}

// Index return node index
func (n *Node) Index() int {
	return n.index
}

// Dag struct
type Dag struct {
	nodes  map[interface{}]*Node
	index  int
	indexs map[int]interface{}
}

// ToProto converts domain Dag into proto Dag
func (dag *Dag) ToProto() (proto.Message, error) {

	nodes := make([]*dagpb.Node, len(dag.nodes))

	for idx, key := range dag.indexs {
		v, ok := dag.nodes[key]
		if !ok {
			return nil, ErrInvalidDagToProto
		}

		node := new(dagpb.Node)
		node.Index = int32(v.index)
		//node.Key = v.Key.(string)
		node.Children = make([]int32, len(v.children))
		for i, child := range v.children {
			node.Children[i] = int32(child.index)
		}

		nodes[idx] = node
	}

	return &dagpb.Dag{
		Nodes: nodes,
	}, nil
}

// FromProto converts proto Dag to domain Dag
func (dag *Dag) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*dagpb.Dag); ok {
		if msg != nil {
			for _, v := range msg.Nodes {
				dag.addNodeWithIndex(int(v.Index), int(v.Index))
			}

			for _, v := range msg.Nodes {
				for _, child := range v.Children {
					dag.AddEdge(int(v.Index), int(child))
				}
			}
			return nil
		}
		return ErrInvalidProtoToDag
	}
	return ErrInvalidProtoToDag
}

// String
func (dag *Dag) String() string {
	msg, err := dag.ToProto()
	if err != nil {
		return string("")
	}
	j, _ := json.Marshal(msg)
	return string(j)
}

// NewDag new dag
func NewDag() *Dag {
	return &Dag{
		nodes:  make(map[interface{}]*Node, 0),
		index:  0,
		indexs: make(map[int]interface{}, 0),
	}
}

// Len Dag len
func (dag *Dag) Len() int {
	return len(dag.nodes)
}

// GetNode get node by key
func (dag *Dag) GetNode(key interface{}) *Node {
	if v, ok := dag.nodes[key]; ok {
		return v
	}
	return nil
}

// GetChildrenNodes get children nodes with key
func (dag *Dag) GetChildrenNodes(key interface{}) []*Node {
	if v, ok := dag.nodes[key]; ok {
		return v.children
	}

	return nil
}

// GetRootNodes get root nodes
func (dag *Dag) GetRootNodes() []*Node {
	nodes := make([]*Node, 0)
	for _, node := range dag.nodes {
		if node.parentCounter == 0 {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetNodes get all nodes
func (dag *Dag) GetNodes() []*Node {
	nodes := make([]*Node, 0)
	for _, node := range dag.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// AddNode add node
func (dag *Dag) AddNode(key interface{}) error {
	if _, ok := dag.nodes[key]; ok {
		return ErrKeyIsExisted
	}

	dag.nodes[key] = NewNode(key, dag.index)
	dag.indexs[dag.index] = key
	dag.index++
	return nil
}

// addNodeWithIndex add node
func (dag *Dag) addNodeWithIndex(key interface{}, index int) error {
	if _, ok := dag.nodes[key]; ok {
		return ErrKeyIsExisted
	}

	dag.nodes[key] = NewNode(key, index)
	dag.indexs[index] = key
	dag.index = index
	return nil
}

// AddEdge add edge fromKey toKey
func (dag *Dag) AddEdge(fromKey, toKey interface{}) error {
	var from, to *Node
	var ok bool

	if from, ok = dag.nodes[fromKey]; !ok {
		return ErrKeyNotFound
	}

	if to, ok = dag.nodes[toKey]; !ok {
		return ErrKeyNotFound
	}

	for _, childNode := range from.children {
		if childNode == to {
			return ErrKeyIsExisted
		}
	}

	dag.nodes[toKey].parentCounter++
	dag.nodes[fromKey].children = append(from.children, to)

	return nil
}

//IsCirclular a->b-c->a
func (dag *Dag) IsCirclular() bool {

	visited := make(map[interface{}]int, len(dag.nodes))
	rootNodes := make(map[interface{}]*Node)
	for key, node := range dag.nodes {
		visited[key] = 0
		rootNodes[key] = node
	}

	for _, node := range rootNodes {
		if dag.hasCirclularDep(node, visited) {
			return true
		}
	}

	for _, count := range visited {
		if count == 0 {
			return true
		}
	}
	return false
}

func (dag *Dag) hasCirclularDep(current *Node, visited map[interface{}]int) bool {

	visited[current.key] = 1
	for _, child := range current.children {
		if visited[child.key] == 1 {
			return true
		}

		if dag.hasCirclularDep(child, visited) {
			return true
		}
	}
	visited[current.key] = 2
	return false
}
