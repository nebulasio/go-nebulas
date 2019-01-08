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

import (
	"os"
	//"os/exec"
	"syscall"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/state"

	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// NebulasVM type of NebulasVM
type NebulasVM struct{
	listenAddr string
}

// NewNebulasVM create new NebulasVM
func NewNebulasVM() core.NVM {
	return &NebulasVM{}
}

func (nvm *NebulasVM) GetNVMListenAddr() string {
	return nvm.listenAddr
}

// Start engine process
func (nvm *NebulasVM) StartNebulasVM(nvmPath string, listenAddr string) (int, error) {

	/*
	cmd := exec.Command(nvmPath, listenAddr)

	err := cmd.Start()
	if err != nil {
		err = errors.New("Failed to start NVM process")
		return 0, err
	}

	pid := cmd.Process.Pid

	*/

	logging.CLog().Info("Started NVM process with port: ", listenAddr)

	pid := 37373		// for debugging purpose

	nvm.listenAddr = listenAddr
	
	return pid, nil
}

// Stop engine process
func (nvm *NebulasVM) StopNebulasVM(enginePid int) error {

	proc, err := os.FindProcess(enginePid)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to find nvm process")
		return err
	}

	err = proc.Kill()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to stop nvm process")
		return err
	}

	logging.CLog().Info("Stopping NVM process")

	return nil
}

// Check if V8 is running
func (nvm *NebulasVM) CheckV8ServerRunning(enginePid int) bool {
	
	proc, err := os.FindProcess(enginePid)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to find nvm process")
		return false
	}
	
	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to ping nvm process")
		return false
	}

	return true
}




//==================== V8 specific =====================

// CreateEngine start engine
func (nvm *NebulasVM) CreateEngine(block *core.Block, tx *core.Transaction, contract state.Account, state core.WorldState) (core.SmartContractEngine, error) {

	ctx, err := NewContext(block, tx, contract, state)
	if err != nil {
		return nil, err
	}
	return NewV8Engine(ctx), nil
}

/*
//func (nvm *NebulasVM) DeployAndInit(block *core.Block, tx *core.Transaction, contract state.Account, state core.WorldState, 
func (nvm *NebulasVM) DeployAndInit(engine core.SmartContractEngine, config *core.NVMConfig, listenAddr string) (core.NVMExeResponse, error) {

	res := core.NVMExeResponse{}

	config.SetFunctionName("init")
	result, exeErr := engine.RunScriptSource(config, listenAddr)
	res.Result = result
	res.GasCount = engine.actualCountOfExecutionInstructions

	return res, exeErr
}

func (nvm *NebulasVM) Call(engine core.SmartContractEngine, config *core.NVMConfig, listenAddr string) (core.NVMExeResponse, error) {

	res := core.NVMExeResponse{}

	if core.PublicFuncNameChecker.MatchString(config.FunctionName) == false {
		logging.VLog().Debugf("Invalid function: %v", config.FunctionName)
		return res, ErrDisallowCallNotStandardFunction
	}
	if strings.EqualFold("init", config.FunctionName) == true {
		return res, ErrDisallowCallPrivateFunction
	}

	result, exeErr := engine.RunScriptSource(config, listenAddr)
	res.Result = result
	res.GasCount = engine.actualCountOfExecutionInstructions

	return res, exeErr
}


func (nvm *NebulasVM) CheckConfig(config *core.NVMConfig) error {
	
	limitsOfExeInstruction := config.LimitsExeInstruction
	limitsTotalMemSize := config.LimitsTotalMemorySize

	if limitsOfExeInstruction == 0 || limitsTotalMemSize == 0 {
		logging.VLog().Debugf("Limit config is empty, limitsofexeinstructions: %v, limitsoftotalmemsize: %d", limitsOfExeInstruction, limitsTotalMemSize)
		return ErrLimitHasEmpty
	}

	if limitsTotalMemSize > 0 && limitsTotalMemSize < 6000000 {
		logging.VLog().Debugf("V8 needs at least 6M (6000000) heap memory, your limitsOfTotalMemSize (%d) is too low.", limitsTotalMemSize)
		return ErrSetMemorySmall
	}

	logging.VLog().WithFields(logrus.Fields{
		"limits_of_executed_instructions": limitsOfExeInstruction,
		"limits_of_total_memory_size": limitsTotalMemSize,
	}).Debug("Set execution limits")

	return nil
}
*/