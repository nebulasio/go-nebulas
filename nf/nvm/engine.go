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
	"errors"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/state"
)

// Error Types
var (
	ErrEngineRepeatedStart = errors.New("engine repeated start")
	ErrEngineNotStart      = errors.New("engine not start")
)

// NebulasVM type of NebulasVM
type NebulasVM struct {
	engine *V8Engine
}

// NewNebulasVM create new NebulasVM
func NewNebulasVM() core.Engine {
	nvm := &NebulasVM{}
	return nvm
}

// CreateEngine start engine
func (nvm *NebulasVM) CreateEngine(block *core.Block, tx *core.Transaction, owner, contract state.Account, state state.AccountState) error {
	if nvm.engine != nil {
		return ErrEngineRepeatedStart
	}

	ctx := &Context{
		block:    block,
		tx:       tx,
		owner:    owner,
		contract: contract,
		state:    state,
	}
	nvm.engine = NewV8Engine(ctx)
	return nil
}

// SetEngineExecutionLimits set limits of execution instructions
func (nvm *NebulasVM) SetEngineExecutionLimits(limitsOfExecutionInstructions uint64) error {
	if nvm.engine == nil {
		return ErrEngineNotStart
	}
	nvm.engine.SetExecutionLimits(limitsOfExecutionInstructions, DefaultLimitsOfTotalMemorySize)
	return nil
}

// DeployAndInitEngine deploy and init source
func (nvm *NebulasVM) DeployAndInitEngine(source, sourceType, args string) (string, error) {
	if nvm.engine == nil {
		return "", ErrEngineNotStart
	}
	return nvm.engine.DeployAndInit(source, sourceType, args)
}

// CallEngine run source function
func (nvm *NebulasVM) CallEngine(source, sourceType, function, args string) (string, error) {
	if nvm.engine == nil {
		return "", ErrEngineNotStart
	}
	return nvm.engine.Call(source, sourceType, function, args)
}

// ExecutionInstructions returns instructions count
func (nvm *NebulasVM) ExecutionInstructions() (uint64, error) {
	if nvm.engine == nil {
		return 0, ErrEngineNotStart
	}
	return nvm.engine.ExecutionInstructions(), nil
}

// DisposeEngine dispose engine
func (nvm *NebulasVM) DisposeEngine() {
	if nvm.engine != nil {
		nvm.engine.Dispose()
		nvm.engine = nil
	}
}

// Clone clone a new engine
func (nvm *NebulasVM) Clone() core.Engine {
	n := &NebulasVM{}
	return n
}
