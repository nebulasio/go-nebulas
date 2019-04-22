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
#include <native/nipc_interface.h>

void NbreVersionFunc_cgo(int isc, void *holder, uint32_t major, uint32_t minor,uint32_t patch);
void NbreIrListFunc_cgo(int isc, void *holder, const char *ir_name_list);
void NbreIrVersionsFunc_cgo(int isc, void *holder, const char *ir_versions);
void NbreNrHandleFunc_cgo(int isc, void *holder, const char *nr_handle);
void NbreNrResultByhandleFunc_cgo(int isc, void *holder, const char *nr_result);
void NbreNrResultByHeightFunc_cgo(int isc, void *holder, const char *nr_result);
void NbreNrSumFunc_cgo(int isc, void *holder, const char *nr_sum);
void NbreDipRewardFunc_cgo(int isc, void *holder, const char *dip_reward);
*/
import "C"
import (
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"

	"unsafe"

	"path/filepath"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/util"
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
	result interface{}
	err    error
	done   chan bool
}

// Nbre type of Nbre
type Nbre struct {
	neb Neblet

	libHeight uint64

	quitCh chan int
}

// NewNbre create new Nbre
func NewNbre(neb Neblet) core.Nbre {
	nbreOnce.Do(func() {
		InitializeNbre()
	})
	return &Nbre{
		neb:       neb,
		libHeight: 2, //lib start from 2
		quitCh:    make(chan int, 1),
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

	if n.neb.BlockChain().LIB() != nil {
		n.libHeight = n.neb.BlockChain().LIB().Height()
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
	if err := util.CreateDirIfNotExist(rootDir); err != nil {
		return err
	}

	if len(conf.LogDir) > 0 {
		logDir = conf.LogDir
	}
	if logDir, err = filepath.Abs(logDir); err != nil {
		return err
	}

	if err := util.CreateDirIfNotExist(logDir); err != nil {
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

	cAdminAddr := C.CString(n.neb.Config().Nbre.AdminAddress)
	defer C.free(unsafe.Pointer(cAdminAddr))

	cIpcListen := C.CString(n.neb.Config().Nbre.IpcListen)
	defer C.free(unsafe.Pointer(cIpcListen))

	p := C.nbre_params_t{
		m_nbre_root_dir:     cRootDir,
		m_nbre_exe_name:     cNbrePath,
		m_neb_db_dir:        cDataDir,
		m_nbre_db_dir:       cNbreDataDir,
		m_nbre_log_dir:      cLogDir,
		m_admin_pub_addr:    cAdminAddr,
		m_nbre_start_height: C.uint64_t(n.neb.Config().Nbre.StartHeight),
		m_nipc_listen:       cIpcListen,
		m_nipc_port:         C.uint16_t(n.neb.Config().Nbre.IpcPort),
	}

	cResult := C.start_nbre_ipc(p)
	if int(cResult) != 0 {
		logging.VLog().WithFields(logrus.Fields{
			"data":   nbreDataDir,
			"nbre":   nbrePath,
			"result": int(cResult),
		}).Error("Failed to start nbre.")
		return ErrNbreStartFailed
	}

	logging.CLog().WithFields(logrus.Fields{
		"data":  nbreDataDir,
		"nbre":  nbrePath,
		"admin": n.neb.Config().Nbre.AdminAddress,
	}).Info("Started nbre.")

	go n.loop()
	return nil
}

func (n *Nbre) loop() {

	logging.CLog().Info("started nbre loop.")

	timerChan := time.NewTicker(time.Second * 15).C
	for {
		select {
		case <-n.quitCh:
			logging.CLog().Info("Stopped nbre ir loop.")
			return
		case <-timerChan:
			n.checkIRUpdate()
		}
	}
}

// checkIRUpdate check lib block for ir transactions packaged.
// If ir transactions are missed, nbre looks for database completion
func (n *Nbre) checkIRUpdate() {
	lib := n.neb.BlockChain().LIB()
	if lib == nil || lib.Height() < n.neb.Config().Nbre.StartHeight {
		return
	}

	for lib.Height() >= n.libHeight {
		block := n.neb.BlockChain().GetBlockOnCanonicalChainByHeight(n.libHeight)
		txs := []*core.Transaction{}
		for _, tx := range block.Transactions() {
			if tx.Type() == core.TxPayloadProtocolType {
				txs = append(txs, tx)
			}
		}

		handle := unsafe.Pointer(uintptr(block.Height()))
		//prepare for tx send
		C.ipc_nbre_ir_transactions_create(handle, C.uint64_t(block.Height()))

		if len(txs) > 0 {
			//append tx data
			for _, tx := range txs {
				pbTx, err := tx.ToProto()
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"tx":  tx,
						"err": err,
					}).Error("Failed to convert the ir tx to proto data.")
					return
				}
				bytes, err := proto.Marshal(pbTx)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"tx":  tx,
						"err": err,
					}).Error("Failed to marshal the ir tx.")
					return
				}
				cBytes := (*C.char)(unsafe.Pointer(&bytes[0]))
				C.ipc_nbre_ir_transactions_append(handle, C.uint64_t(block.Height()), cBytes, C.int32_t(len(bytes)))
			}
		}

		// commit tx send
		C.ipc_nbre_ir_transactions_send(handle, C.uint64_t(block.Height()))

		logging.VLog().WithFields(logrus.Fields{
			"height":   block.Height(),
			"ir count": len(txs),
		}).Debug("Update ir block.")

		n.libHeight++
	}
}

// InitializeNbre initialize nbre
func InitializeNbre() {
	C.set_recv_nbre_version_callback((C.nbre_version_callback_t)(unsafe.Pointer(C.NbreVersionFunc_cgo)))
	C.set_recv_nbre_ir_list_callback((C.nbre_ir_list_callback_t)(unsafe.Pointer(C.NbreIrListFunc_cgo)))
	C.set_recv_nbre_ir_versions_callback((C.nbre_ir_versions_callback_t)(unsafe.Pointer(C.NbreIrVersionsFunc_cgo)))
	C.set_recv_nbre_nr_handle_callback((C.nbre_nr_handle_callback_t)(unsafe.Pointer(C.NbreNrHandleFunc_cgo)))
	C.set_recv_nbre_nr_result_by_handle_callback((C.nbre_nr_result_by_handle_callback_t)(unsafe.Pointer(C.NbreNrResultByhandleFunc_cgo)))
	C.set_recv_nbre_nr_result_by_height_callback((C.nbre_nr_result_by_height_callback_t)(unsafe.Pointer(C.NbreNrResultByHeightFunc_cgo)))
	C.set_recv_nbre_nr_sum_callback((C.nbre_nr_sum_callback_t)(unsafe.Pointer(C.NbreNrSumFunc_cgo)))
	C.set_recv_nbre_dip_reward_callback((C.nbre_dip_reward_callback_t)(unsafe.Pointer(C.NbreDipRewardFunc_cgo)))
}

// Execute execute command
func (n *Nbre) Execute(command string, args ...interface{}) (interface{}, error) {
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

	logging.VLog().WithFields(logrus.Fields{
		"id":      handler.id,
		"command": command,
		"args":    args,
	}).Debug("run nbre command")

	go func() {
		// handle nbre command
		n.handleNbreCommand(handler, command, args...)
	}()

	select {
	case <-handler.done:
		// wait for C func returns.
		deleteHandler(handler)

	case <-time.After(ExecutionTimeoutSeconds * time.Second):
		handler.err = ErrNebCallbackTimeout
		deleteHandler(handler)
	}

	logging.VLog().WithFields(logrus.Fields{
		"id":      handler.id,
		"command": command,
		"params":  args,
		"result":  handler.result,
		"error":   handler.err,
	}).Debug("nbre command response")
	return handler.result, handler.err
}

func deleteHandler(handler *handler) {
	nbreLock.Lock()
	defer nbreLock.Unlock()
	handler.done = nil
	delete(nbreHandlers, handler.id)
}

func (n *Nbre) handleNbreCommand(handler *handler, command string, args ...interface{}) {
	handlerId := uint64(0)
	if handler != nil {
		handlerId = handler.id
	}

	switch command {
	case CommandVersion:
		height := args[0].(uint64)
		C.ipc_nbre_version(unsafe.Pointer(uintptr(handlerId)), C.uint64_t(height))
	case CommandIRList:
		C.ipc_nbre_ir_list(unsafe.Pointer(uintptr(handlerId)))
	case CommandIRVersions:
		irName := args[0].(string)
		cIrName := C.CString(irName)
		defer C.free(unsafe.Pointer(cIrName))
		C.ipc_nbre_ir_versions(unsafe.Pointer(uintptr(handlerId)), cIrName)
	case CommandNRHandler:
		start := args[0].(uint64)
		end := args[1].(uint64)
		version := args[2].(uint64)
		C.ipc_nbre_nr_handle(unsafe.Pointer(uintptr(handlerId)), C.uint64_t(start), C.uint64_t(end), C.uint64_t(version))
	case CommandNRListByHandle:
		handle := args[0].(string)
		cHandle := C.CString(handle)
		defer C.free(unsafe.Pointer(cHandle))
		C.ipc_nbre_nr_result_by_handle(unsafe.Pointer(uintptr(handlerId)), cHandle)
	case CommandNRListByHeight:
		height := args[0].(uint64)
		C.ipc_nbre_nr_result_by_height(unsafe.Pointer(uintptr(handlerId)), C.uint64_t(height))
	case CommandNRSum:
		height := args[0].(uint64)
		C.ipc_nbre_nr_sum(unsafe.Pointer(uintptr(handlerId)), C.uint64_t(height))
	case CommandDIPList:
		height := args[0].(uint64)
		version := args[1].(uint64)
		C.ipc_nbre_dip_reward(unsafe.Pointer(uintptr(handlerId)), C.uint64_t(height), C.uint64_t(version))
	default:
		if handler != nil {
			handler.result = nil
			handler.err = ErrCommandNotFound
			if handler.done != nil {
				handler.done <- true
			}
		}
	}
}

func getNbreHandler(id uint64) (*handler, error) {
	nbreLock.Lock()
	defer nbreLock.Unlock()

	if handler, ok := nbreHandlers[id]; ok {
		return handler, nil
	} else {
		return nil, ErrHandlerNotFound
	}
}

func nbreHandled(code C.int, holder unsafe.Pointer, result interface{}, handleErr error) {
	handlerId := uint64(uintptr(holder))
	handler, err := getNbreHandler(handlerId)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handlerId": handlerId,
			"err":       err,
		}).Error("Failed to handle nbre callback")
		return
	}

	switch code {
	case C.ipc_status_succ:
		err = nil
	case C.ipc_status_fail:
		err = ErrNbreCallbackFailed
	case C.ipc_status_timeout:
		err = ErrExecutionTimeout
	case C.ipc_status_exception:
		err = ErrNbreCallbackException
	case C.ipc_status_nbre_not_ready:
		err = ErrNbreCallbackNotReady
	default:
		err = ErrNbreCallbackCodeErr
	}

	if err == nil {
		handler.result = result
		handler.err = handleErr
	} else {
		handler.err = err
	}
	go func() {
		if handler.done != nil {
			handler.done <- true
		}
	}()
}

// Stop stop nbre
func (n *Nbre) Stop() {
	logging.CLog().Info("Stopping Nbre.")

	// stop ir check loop
	n.quitCh <- 1

	select {
	case <-n.shutdown():
		return
	}
}

// Shutdown shutdown nbre
func (n *Nbre) shutdown() chan bool {
	quitCh := make(chan bool, 1)

	go func() {
		C.nbre_ipc_shutdown()
		logging.CLog().Info("Stopped Nbre.")

		quitCh <- true
		return
	}()

	return quitCh
}
