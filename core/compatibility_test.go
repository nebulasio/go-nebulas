// Copyright (C) 2018 go-nebulas authors
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

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringCompare(t *testing.T) {
	type test struct {
		name, a, b string
		cmp        int
	}

	tests := []test{
		{
			name: "1",
			a:    "1.0.1",
			b:    "1.0.2",
			cmp:  -1,
		},
		{
			name: "2",
			a:    "1.0.1",
			b:    "1.0.10",
			cmp:  -1,
		},
		{
			name: "3",
			a:    "1.0.10",
			b:    "1.0.2",
			cmp:  1,
		},
		{
			name: "4",
			a:    "10.0.1",
			b:    "1.0.2",
			cmp:  1,
		},
		{
			name: "5",
			a:    "10.0.1",
			b:    "2.0.2",
			cmp:  1,
		},
		{
			name: "6",
			a:    "10.0.1",
			b:    "10.90.1",
			cmp:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va, _ := parseVersion(tt.a)
			vb, _ := parseVersion(tt.b)
			r := compareVersion(va, vb)
			assert.Equal(t, tt.a, va.String())
			assert.Equal(t, tt.b, vb.String())
			assert.Equal(t, tt.cmp, r, "case "+tt.name+" not equal")
		})
	}
}
