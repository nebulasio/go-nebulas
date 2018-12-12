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

void NbreVersionFunc_cgo(int isc, void *holder, uint32_t major, uint32_t minor,uint32_t patch);
void NbreNrFunc_cgo(int isc, void *holder, const char *nr_result);
*/
import "C"
import (
	"sync"
	"time"

	"unsafe"

	"path/filepath"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	// ExecutionTimeoutSeconds max nbre execution timeout.
	ExecutionTimeoutSeconds = 15
)

// default config path
const (
	defaultRootDir     = "nbre"
	defaultLogDir      = "nbre/logs"
	defaultNbreDataDir = "nbre/nbre.db"
	defaultNbrePath    = "nbre/nbre"
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

//func getCurrPath() string {
//	file, _ := exec.LookPath(os.Args[0])
//	path, _ := filepath.Abs(file)
//	index := strings.LastIndex(path, string(os.PathSeparator))
//	ret := path[:index]
//	return ret
//}

// Start launch the nbre
func (n *Nbre) Start() error {
	if n.neb.Config().Nbre == nil {
		return ErrConfigNotFound
	}
	var (
		rootDir     = defaultRootDir
		logDir      = defaultLogDir
		nbreDataDir = defaultNbreDataDir
		nbrePath    = defaultNbrePath
		err         error
	)
	conf := n.neb.Config().Nbre
	if len(conf.RootDir) > 0 {
		rootDir = conf.RootDir
	}
	if rootDir, err = filepath.Abs(rootDir); err != nil {
		return err
	}

	if len(conf.LogDir) > 0 {
		logDir = conf.LogDir
	}
	if logDir, err = filepath.Abs(logDir); err != nil {
		return err
	}

	if len(conf.DataDir) > 0 {
		nbreDataDir = conf.DataDir
	}
	if nbreDataDir, err = filepath.Abs(nbreDataDir); err != nil {
		return err
	}

	if len(conf.NbrePath) > 0 {
		nbrePath = conf.NbrePath
	}
	if nbrePath, err = filepath.Abs(nbrePath); err != nil {
		return err
	}

	dataDir, err := filepath.Abs(n.neb.Config().Chain.Datadir)
	if err != nil {
		return err
	}

	cRootDir := C.CString(rootDir)
	defer C.free(unsafe.Pointer(cRootDir))
	cLogDir := C.CString(logDir)
	defer C.free(unsafe.Pointer(cLogDir))
	cNbreDataDir := C.CString(nbreDataDir)
	defer C.free(unsafe.Pointer(cNbreDataDir))
	cNbrePath := C.CString(nbrePath)
	defer C.free(unsafe.Pointer(cNbrePath))
	cDataDir := C.CString(dataDir)
	defer C.free(unsafe.Pointer(cDataDir))

	cAddr := C.CString(n.neb.Config().Nbre.AdminAddress)
	defer C.free(unsafe.Pointer(cAddr))

	p := C.nbre_params_t{
		m_nbre_root_dir:  cRootDir,
		m_nbre_exe_name:  cNbrePath,
		m_neb_db_dir:     cDataDir,
		m_nbre_db_dir:    cNbreDataDir,
		m_nbre_log_dir:   cLogDir,
		m_admin_pub_addr: cAddr,
	}
	//p.m_nbre_root_dir = cRootDir
	//p.m_nbre_exe_name = cNbrePath
	//p.m_neb_db_dir = cDataDir
	//p.m_nbre_db_dir = cNbreDataDir
	//p.m_nbre_log_dir = cLogDir
	//p.m_admin_pub_addr = cAddr

	logging.CLog().WithFields(logrus.Fields{
		"data":  nbreDataDir,
		"nbre":  nbrePath,
		"admin": n.neb.Config().Nbre.AdminAddress,
	}).Info("Started nbre.")

	cResult := C.start_nbre_ipc(p)
	if int(cResult) != 0 {
		return ErrNbreStartFailed
	}
	return nil
}

// InitializeNbre initialize nbre
func InitializeNbre() {
	C.set_recv_nbre_version_callback((C.nbre_version_callback_t)(unsafe.Pointer(C.NbreVersionFunc_cgo)))
	C.set_recv_nbre_nr_callback((C.nbre_nr_callback_t)(unsafe.Pointer(C.NbreNrFunc_cgo)))
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
	case CommandNRList:
		pStr := &Params{}
		pStr.FromBytes(params)
		C.ipc_nbre_nr(unsafe.Pointer(uintptr(handlerId)), C.uint64_t(pStr.StartBlock), C.uint64_t(pStr.EndBlock), C.uint64_t(pStr.Version))
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
