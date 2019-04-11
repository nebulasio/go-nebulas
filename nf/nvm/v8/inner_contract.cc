// Copyright (C) 2017-2019 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.

#include "v8_util.h"

const uint32_t NVMEngine::CreateInnerContractEngine(std::string scriptType, std::string innerContractSrc){

    if(this->m_inner_engines == nullptr)
        this->m_inner_engines = std::unique_ptr<std::vector<V8Engine*>>(new std::vector<V8Engine*>());

    V8Engine* engine = CreateEngine();
    ReadMemoryStatistics(this->engine);

    engine->limits_of_executed_instructions = configBundle.limits_exe_instruction();
    engine->limits_of_total_memory_size = configBundle.limits_total_mem_size();

    // transpile script and inject tracing code if necessary
    int runnableSourceResult = this->GetRunnableSourceCode(scriptType, innerContractSrc);
    if(runnableSourceResult != 0){
      LogErrorf("Failed to get runnable source code");
      return runnableSourceResult;
    }

    AddModule(this->engine, moduleID.c_str(), this->m_traceable_src.c_str(), this->m_traceale_src_line_offset);

    // set limitation for the new engine
    this->m_inner_engines->push_back(engine);
}

// start one level of inner contract call, check the resource usage firstly
const void NVMEngine::StartInnerContractCall(){

    //TODO: check left resource in V8 process
	remainInstruction, remainMem := engine.GetNVMLeftResources()
	if remainInstruction <= uint64(InnerContractGasBase) {
		logging.VLog().WithFields(logrus.Fields{
			"remainInstruction": remainInstruction,
			"mem":               remainMem,
			"err":               ErrInnerInsufficientGas.Error(),
		}).Error("failed to prepare create nvm")
		setHeadErrAndLog(engine, index, ErrInsufficientGas, "null", false)
		return emptyRes, gasCnt, false
	} else {
		remainInstruction -= InnerContractGasBase
	}

	if remainMem <= 0 {
		logging.VLog().WithFields(logrus.Fields{
			"remainInstruction": remainInstruction,
			"mem":               remainMem,
			"err":               ErrInnerInsufficientMem.Error(),
		}).Error("failed to prepare create nvm")
		setHeadErrAndLog(engine, index, ErrExceedMemoryLimits, "null", false)
		return emptyRes, gasCnt, false
	}

    logging.VLog().Debugf("begin create New V8,intance:%v, mem:%v", remainInstruction, remainMem)


}