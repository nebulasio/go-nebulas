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

	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

const (

	// EventNameSpaceContract the topic of contract.
	EventNameSpaceContract = "chain.contract"
)

// EventTriggerFunc export EventTriggerFunc
//export EventTriggerFunc
func EventTriggerFunc(handler unsafe.Pointer, topic, data *C.char) {
	gTopic := C.GoString(topic)
	gData := C.GoString(data)

	e := getEngineByEngineHandler(handler)
	if e == nil {
		log.WithFields(log.Fields{
			"category": 0, // ChainEventCategory.
			"topic":    gTopic,
			"data":     gData,
		}).Error("Event.Trigger delegate handler does not found.")
		return
	}

	log.WithFields(log.Fields{
		"category": 0, // ChainEventCategory.
		"topic":    gTopic,
		"data":     gData,
	}).Info("Event triggered from V8 engine.")

	txHash, _ := byteutils.FromHex(e.ctx.tx.Hash)
	contractTopic := EventNameSpaceContract + "." + gTopic
	e.ctx.block.RecordEvent(txHash, contractTopic, gData)
}
