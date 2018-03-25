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

package nvm

import "C"
import (
	"unsafe"

	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	// EventBaseGasCount the gas count of a new event
	EventBaseGasCount = 20
)

// EventTriggerFunc export EventTriggerFunc
//export EventTriggerFunc
func EventTriggerFunc(handler unsafe.Pointer, topic, data *C.char, gasCnt *C.size_t) {
	gTopic := C.GoString(topic)
	gData := C.GoString(data)

	e := getEngineByEngineHandler(handler)
	if e == nil {
		logging.VLog().WithFields(logrus.Fields{
			"category": 0, // ChainEventCategory.
			"topic":    gTopic,
			"data":     gData,
		}).Error("Event.Trigger delegate handler does not found.")
		return
	}

	// calculate Gas.
	*gasCnt = C.size_t(EventBaseGasCount + len(gTopic) + len(gData))

	contractTopic := EventNameSpaceContract + "." + gTopic
	event := &state.Event{Topic: contractTopic, Data: gData}
	e.ctx.state.RecordEvent(e.ctx.tx.Hash(), event)
}
