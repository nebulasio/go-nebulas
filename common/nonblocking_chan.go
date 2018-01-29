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

package common

import (
	"time"
)

type NonBlockingChan struct {
	innerChan chan interface{}
}

func NewNonBlockingChan(cap int) *NonBlockingChan {
	return &NonBlockingChan{
		innerChan: make(chan interface{}, cap),
	}
}

func (nbCh *NonBlockingChan) Send(value interface{}) bool {
	select {
	case nbCh.innerChan <- value:
		return true
	default:
		return false
	}
}

func (nbCh *NonBlockingChan) Recv() (interface{}, bool) {
	select {
	case value := <-nbCh.innerChan:
		return value, true
	default:
		return nil, false
	}
}

func (nbCh *NonBlockingChan) SendWithDeadline(value interface{}, deadline time.Duration) bool {
	if deadline == 0 {
		return nbCh.Send(value)
	}

	select {
	case nbCh.innerChan <- value:
		return true
	case <-time.NewTicker(deadline).C:
		return false
	}
}

func (nbCh *NonBlockingChan) RecvWithDeadline(deadline time.Duration) (interface{}, bool) {
	if deadline == 0 {
		return nbCh.Recv()
	}

	select {
	case value := <-nbCh.innerChan:
		return value, true
	case <-time.NewTicker(deadline).C:
		return nil, false
	}
}
