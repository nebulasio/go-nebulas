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

// Cmp function, a < b -> -1, a == b -> 0, a > b -> 1
type Cmp func(a interface{}, b interface{}) int

// Slice is a sorted array
type Slice struct {
	content []interface{}
	cmp     Cmp
}

// NewSlice return a new slice
func NewSlice(cmp Cmp) *Slice {
	return &Slice{
		cmp: cmp,
	}
}

// Push a new value into slice
func (s *Slice) Push(val interface{}) {
	if len(s.content) == 0 {
		s.content = append(s.content, val)
		return
	}

	start, end := 0, len(s.content)-1
	result, mid := 0, 0
	for start <= end {
		mid = (start + end) / 2
		result = s.cmp(s.content[mid], val)
		if result > 0 {
			end = mid - 1
		} else if result < 0 {
			start = mid + 1
		} else {
			break
		}
	}
	content := []interface{}{val}
	if result > 0 {
		content = append(content, s.content[mid:]...)
		content = append(s.content[0:mid], content...)
	} else {
		content = append(content, s.content[mid+1:]...)
		content = append(s.content[0:mid+1], content...)

	}
	s.content = content
}

// PopLeft pop out the min value
func (s *Slice) PopLeft() interface{} {
	if s.Len() > 0 {
		val := s.content[0]
		s.content = s.content[1:]
		return val
	}
	return nil
}

// PopRight pop out the max value
func (s *Slice) PopRight() interface{} {
	if s.Len() > 0 {
		val := s.content[s.Len()-1]
		s.content = s.content[:s.Len()-1]
		return val
	}
	return nil
}

// Del the given value
func (s *Slice) Del(val interface{}) {
	for k, v := range s.content {
		if v == val {
			var content []interface{}
			content = append(content, s.content[k+1:]...)
			content = append(s.content[0:k], content...)
			s.content = content
			return
		}
	}
}

// Index return the value at the given value
func (s *Slice) Index(index int) interface{} {
	if s.Len() > index {
		return s.content[index]
	}
	return nil
}

// Len return the length of slice
func (s *Slice) Len() int {
	return len(s.content)
}

// Left return the min value, not pop out
func (s *Slice) Left() interface{} {
	if s.Len() > 0 {
		return s.content[0]
	}
	return nil
}

// Right return the max value, not pop out
func (s *Slice) Right() interface{} {
	if s.Len() > 0 {
		return s.content[len(s.content)-1]
	}
	return nil
}
