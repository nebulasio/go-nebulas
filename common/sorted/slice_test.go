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

package sorted

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testCmp(a interface{}, b interface{}) int {
	ai := a.(int)
	bi := b.(int)
	if ai < bi {
		return -1
	} else if ai > bi {
		return 1
	} else {
		return 0
	}
}

func TestSlice(t *testing.T) {
	slice := NewSlice(testCmp)
	slice.Push(3)
	slice.Push(2)
	slice.Push(4)
	assert.Equal(t, slice.Left(), 2)
	assert.Equal(t, slice.Right(), 4)
	assert.Equal(t, slice.PopLeft(), 2)
	slice.Del(4)
	slice.Push(1)
	assert.Equal(t, slice.Right(), 3)
	assert.Equal(t, slice.Left(), 1)
}
