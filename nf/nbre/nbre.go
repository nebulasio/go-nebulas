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

package nbre

/*

#include <stdlib.h>
#include <native/ipc_interface.h>

void IpcNbreVersionFunc_cgo(void *holder, uint32_t major, uint32_t minor,uint32_t patch);
*/
import "C"

//#cgo LDFLAGS: -L${SRCDIR}/native -lnbre

import (
	"sync"
	"time"
	"unsafe"

	"github.com/nebulasio/go-nebulas/core"
)

const (
	// ExecutionTimeoutSeconds max nbre execution timeout.
	ExecutionTimeoutSeconds = 15
)

// nbre private data
var (
	nbreOnce     = sync.Once{}
	handlerIdx   = uint64(0)
	nbreHandlers = make(map[uint64]*handler, 1024)
	nbreLock     = sync.RWMutex{}
)

type handler struct {
	id     uint64
	result []byte
	err    error
	done   chan bool
}

// Nbre type of Nbre
type Nbre struct {
	neb Neblet
}

// NewNbre create new Nbre
func NewNbre(neb Neblet) core.Nbre {
	nbreOnce.Do(func() {
		InitializeNbre()
	})
	return &Nbre{
		neb: neb,
	}
}

// Start launch the nbre
func (n *Nbre) Start() error {
	// TODO(larry): add to config
	root := ""
	path := ""
	cRoot := C.CString(root)
	defer C.free(unsafe.Pointer(cRoot))
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	cResult := C.start_nbre_ipc(cRoot, cPath)
	if int(cResult) != 0 {
		return ErrNbreStartFailed
	}
	return nil
}

// InitializeNbre initialize nbre
func InitializeNbre() {
	C.set_recv_nbre_version_callback(C.IpcNbreVersionFunc_cgo)
}

// Execute execute command
func (n *Nbre) Execute(command string, params []byte) ([]byte, error) {
	handlerIdx++
	handler := &handler{
		id:     handlerIdx,
		done:   make(chan bool, 1),
		err:    nil,
		result: nil,
	}

	(func() {
		nbreLock.Lock()
		defer nbreLock.Unlock()
		nbreHandlers[handler.id] = handler
	})()

	go func() {
		// handle nbre command
		n.handleNbreCommand(handler, command, params)
	}()

	select {
	case <-handler.done:
		// wait for C.RunScriptSource() returns.
		nbreLock.Lock()
		delete(nbreHandlers, handler.id)
		nbreLock.Unlock()
	case <-time.After(ExecutionTimeoutSeconds * time.Second):
		handler.err = ErrExecutionTimeout
		select {
		case <-handler.done:
		}
	}
	return handler.result, handler.err
}

func (n *Nbre) handleNbreCommand(handler *handler, command string, params []byte) {
	height := n.neb.BlockChain().TailBlock().Height()
	switch command {
	case CommandVersion:
		C.ipc_nbre_version(unsafe.Pointer(&handler.id), C.uint32_t(height))
	default:
		handler.err = ErrCommandNotFound
		handler.done <- true
	}
}

func getNbreHander(id uint64) (*handler, error) {
	nbreLock.RLock()
	handler := nbreHandlers[id]
	nbreLock.RUnlock()

	if handler == nil {
		return nil, ErrHandlerNotFound
	}
	return handler, nil
}

func nbreHandled(handler *handler, result []byte, err error) {
	if err == nil {
		handler.result = result
	}
	handler.err = err
	handler.done <- true
}

// Shutdown shutdown nbre
func (n *Nbre) Shutdown() error {
	C.nbre_ipc_shutdown()
	return nil
}
