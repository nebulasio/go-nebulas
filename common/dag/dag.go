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

// Vertex struct
type Vertex struct {
	Key           string
	Value         interface{}
	Children      []*Vertex
	ParentCounter int
}

var (
	ErrKeyNotFound  = errors.New("not found")
	ErrKeyIsExisted = errors.New("already existed")
)

// NewVertex new vertex
func NewVertex(key string, value interface{}) *Vertex {
	return &Vertex{
		Key:           key,
		Value:         value,
		ParentCounter: 0,
		Children:      make([]*Vertex, 0),
	}
}

// Dag struct
type Dag struct {
	Vertices map[string]*Vertex
}

// ToProto converts domain Dag into proto Dag
func (dag *Dag) ToProto() (proto.Message, error) {
	vertices := make([]*dagpb.Vertex, len(dag.Vertices))

	idx := 0
	for _, v := range dag.Vertices {
		vertex := new(dagpb.Vertex)
		vertex.Key = v.Key
		vertex.Children = make([]string, len(v.Children))
		for i, child := range v.Children {
			vertex.Children[i] = child.Key
		}

		vertices[idx] = vertex
		idx++
	}

	return &dagpb.Dag{
		Vertices: vertices,
	}, nil
}

// FromProto converts proto Dag to domain Dag
func (dag *Dag) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*dagpb.Dag); ok {
		//dag.cont
		dag.Vertices = make(map[string]*Vertex, len(msg.Vertices))

		for _, v := range msg.Vertices {
			dag.AddVertex(v.Key, nil)
		}

		for _, v := range msg.Vertices {
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
		Vertices: make(map[string]*Vertex, 0),
	}
}

// Len Dag len
func (dag *Dag) Len() int {
	return len(dag.Vertices)
}

// GetVertex get vertex with key
func (dag *Dag) GetVertex(key string) *Vertex {
	if v, ok := dag.Vertices[key]; ok {
		return v
	}
	return nil
}

// GetChildrenVertices get children vertex with key
func (dag *Dag) GetChildrenVertices(key string) []*Vertex {
	if v, ok := dag.Vertices[key]; ok {
		return v.Children
	}
	return nil
}

// GetRootVertices get root vertices
func (dag *Dag) GetRootVertices() []*Vertex {
	vertices := make([]*Vertex, 0)
	for _, vertex := range dag.Vertices {
		if vertex.ParentCounter == 0 {
			vertices = append(vertices, vertex)
		}
	}
	return vertices
}

// GetVertices get all vertices
func (dag *Dag) GetVertices() []*Vertex {
	vertices := make([]*Vertex, 0)
	for _, vertex := range dag.Vertices {
		vertices = append(vertices, vertex)
	}
	return vertices
}

// AddVertex add vertex
func (dag *Dag) AddVertex(key string, value interface{}) error {
	if _, ok := dag.Vertices[key]; ok {
		return ErrKeyIsExisted
	}

	dag.Vertices[key] = NewVertex(key, value)
	return nil
}

// AddEdge add edge fromKey toKey
func (dag *Dag) AddEdge(fromKey, toKey string) error {
	var from, to *Vertex
	var ok bool

	if from, ok = dag.Vertices[fromKey]; !ok {
		return ErrKeyNotFound
	}

	if to, ok = dag.Vertices[toKey]; !ok {
		return ErrKeyNotFound
	}

	for _, childVertex := range from.Children {
		if childVertex == to {
			return ErrKeyIsExisted //todo
		}
	}

	dag.Vertices[toKey].ParentCounter++
	from.Children = append(from.Children, to)
	return nil
}
