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
	"reflect"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestDag_AddVertex(t *testing.T) {
	dag := NewDag()

	dag.AddVertex("1", nil)
	dag.AddVertex("2", nil)
	dag.AddVertex("3", nil)
	dag.AddVertex("4", nil)

	// Add the edges (Note that given vertices must exist before adding an
	// edge between them)
	dag.AddEdge("1", "2")
	dag.AddEdge("1", "3")
	dag.AddEdge("2", "4")
	dag.AddEdge("3", "4")

	vertex := dag.GetVertex("1")

	assert.Equal(t, "1", vertex.Key)
	assert.Equal(t, 0, vertex.ParentCounter)

	vertex4 := dag.GetVertex("4")
	assert.Equal(t, 2, vertex4.ParentCounter)

	dag.AddVertex("5", nil)
	msg, err := dag.ToProto()

	assert.Nil(t, err)

	err = dag.FromProto(msg)

	assert.Nil(t, err)

	vertex1 := dag.GetVertex("5")

	assert.Equal(t, "5", vertex1.Key)
}

func TestNewDag(t *testing.T) {
	tests := []struct {
		name string
		want *Dag
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDag(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDag_FromProto(t *testing.T) {
	type fields struct {
		Vertices map[string]*Vertex
	}
	type args struct {
		msg proto.Message
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dag := &Dag{
				Vertices: tt.fields.Vertices,
			}
			if err := dag.FromProto(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Dag.FromProto() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
