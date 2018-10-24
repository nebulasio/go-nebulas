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
#cgo LDFLAGS: -L${SRCDIR}/native -lnbre_rt

#include <stdlib.h>
#include <native/ipc_interface.h>

void NbreVersionFunc_cgo(void *holder, uint32_t major, uint32_t minor,uint32_t patch);
*/
import "C"
import (
	"sync"
	"time"

	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"unsafe"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
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

func getCurrPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	ret := path[:index]
	return ret
}

// Start launch the nbre
func (n *Nbre) Start() error {
	// TODO(larry): add to config
	root := getCurrPath() + "/nbre/"
	path := root + "bin/nbre"
	cRoot := C.CString(root)
	defer C.free(unsafe.Pointer(cRoot))
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	logging.CLog().WithFields(logrus.Fields{
		"root": root,
		"path": path,
	}).Info("Started nbre.")

	cResult := C.start_nbre_ipc(cRoot, cPath)
	if int(cResult) != 0 {
		return ErrNbreStartFailed
	}
	return nil
}

// InitializeNbre initialize nbre
func InitializeNbre() {
	C.set_recv_nbre_version_callback((C.nbre_version_callback_t)(unsafe.Pointer(C.NbreVersionFunc_cgo)))
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

	//(func() {
	nbreLock.Lock()
	nbreHandlers[handler.id] = handler
	nbreLock.Unlock()
	//})()

	go func() {
		// handle nbre command
		n.handleNbreCommand(handler, command, params)
	}()

	select {
	case <-handler.done:
		// wait for C func returns.
		deleteHandler(handler)

	case <-time.After(ExecutionTimeoutSeconds * time.Second):
		handler.err = ErrExecutionTimeout
		// handler.done <- true
		deleteHandler(handler)
		logging.CLog().WithFields(logrus.Fields{
			"command": command,
			"params":  string(params),
		}).Debug("nbre response timeout.")
	}

	logging.CLog().WithFields(logrus.Fields{
		"command": command,
		"params":  string(params),
		"result":  string(handler.result),
		"error":   handler.err,
	}).Debug("nbre command response")
	return handler.result, handler.err
}

func deleteHandler(handler *handler) {
	nbreLock.Lock()
	defer nbreLock.Unlock()
	delete(nbreHandlers, handler.id)
}

func (n *Nbre) handleNbreCommand(handler *handler, command string, params []byte) {
	height := n.neb.BlockChain().TailBlock().Height()
	handlerId := handler.id

	logging.CLog().WithFields(logrus.Fields{
		"command": command,
		"params":  string(params),
	}).Debug("run nbre command")
	switch command {
	case CommandVersion:
		C.ipc_nbre_version(unsafe.Pointer(uintptr(handlerId)), C.uint64_t(height))
	default:
		handler.err = ErrCommandNotFound
		handler.done <- true
	}
}

func getNbreHandler(id uint64) (*handler, error) {
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
	logging.CLog().Info("Stopped Nbre.")
	return nil
}
