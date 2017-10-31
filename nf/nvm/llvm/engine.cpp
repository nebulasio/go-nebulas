// Copyright (C) 2017 go-nebulas authors
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
//

#include "engine.h"
#include "memory_manager.h"

#include <llvm/ExecutionEngine/ExecutionEngine.h>
#include <llvm/ExecutionEngine/GenericValue.h>
#include <llvm/ExecutionEngine/SectionMemoryManager.h>
#include <llvm/IR/LLVMContext.h>
#include <llvm/IR/LegacyPassManager.h>
#include <llvm/IR/Module.h>
#include <llvm/IRReader/IRReader.h>
#include <llvm/Support/Host.h>
#include <llvm/Support/SourceMgr.h>
#include <llvm/Support/TargetRegistry.h>
#include <llvm/Support/TargetSelect.h>
#include <llvm/Support/raw_ostream.h>
#include <llvm/Transforms/Scalar.h>
#include <llvm/Transforms/Scalar/GVN.h>

#include <stdio.h>
#include <stdlib.h>

using namespace llvm;

extern void LogError(const char *msg);

void Initialize() {
  // Initialization.
  InitializeNativeTarget();
  InitializeNativeTargetAsmParser();
  InitializeNativeTargetAsmPrinter();
  LLVMLinkInMCJIT();
}

void SetTargetAndDataLayout(Module *module) {
  std::string targetTriple = sys::getProcessTriple();
  module->setTargetTriple(targetTriple);

  std::string err;
  auto target = TargetRegistry::lookupTarget(targetTriple, err);

  StringRef cpu = sys::getHostCPUName();

  SubtargetFeatures features;
  StringMap<bool> HostFeatures;
  if (sys::getHostCPUFeatures(HostFeatures))
    for (auto &F : HostFeatures)
      features.AddFeature(F.first(), F.second);
  std::string featureStr = features.getString();

  TargetOptions opt;
  auto rm = Optional<Reloc::Model>();
  auto targetMachine =
      target->createTargetMachine(targetTriple, cpu, featureStr, opt, rm);

  module->setDataLayout(targetMachine->createDataLayout());

  // printf("targetTriple = %s\n", targetTriple.c_str());
  // printf("cpu = %s\n", cpu.begin());
  // printf("featureStr = %s\n", featureStr.c_str());
}

Engine *CreateEngine() {
  // Parse IR.
  LLVMContext *context = new LLVMContext();

  // Create PassManager.
  legacy::PassManager *passMgr = new legacy::PassManager();
  passMgr->add(createConstantPropagationPass());
  passMgr->add(createInstructionCombiningPass());
  passMgr->add(createPromoteMemoryToRegisterPass());
  passMgr->add(createCFGSimplificationPass());
  passMgr->add(createDeadCodeEliminationPass());
  passMgr->add(createGVNPass());

  // Enable MCJIT.
  MemoryManager *rtDyldMM = new MemoryManager();

  // Create Engine Structure.
  Engine *e = static_cast<Engine *>(calloc(1, sizeof(Engine)));
  e->llvm_context = context;
  e->llvm_pass_manager = passMgr;
  e->llvm_builder = NULL;
  e->llvm_mem_manager = rtDyldMM;

  e->llvm_engine = NULL;
  e->llvm_main_module = NULL;
  return e;
}

int AddModuleFile(Engine *e, const char *irPath) {
  LLVMContext *context = static_cast<LLVMContext *>(e->llvm_context);
  legacy::PassManager *passMgr =
      static_cast<legacy::PassManager *>(e->llvm_pass_manager);
  MemoryManager *rtDyldMM = static_cast<MemoryManager *>(e->llvm_mem_manager);

  SMDiagnostic err;

  std::unique_ptr<Module> pModule =
      parseIRFile(StringRef(irPath), err, *context);
  Module *module = pModule.get();
  if (module == nullptr) {
    LogError(err.getMessage().data());
    return 1;
  }

  SetTargetAndDataLayout(module);

  passMgr->run(*module);

  if (false) {
    // TODO: @robin, invalid IR file should be quit.
    LogError("running pass failed.");
    return 1;
  }

  // Create EngineBuilder if not.
  EngineBuilder *builder = static_cast<EngineBuilder *>(e->llvm_builder);
  if (builder == nullptr) {
    std::string errMsg;

    builder = new EngineBuilder(std::move(pModule));
    builder->setErrorStr(&errMsg);
    builder->setEngineKind(EngineKind::JIT);
    builder->setUseOrcMCJITReplacement(false);

    builder->setMCJITMemoryManager(
        std::unique_ptr<RTDyldMemoryManager>(rtDyldMM));

    builder->setOptLevel(CodeGenOpt::Default);

    e->llvm_builder = builder;
    e->llvm_main_module = module;
  }

  // Create ExecutionEngine if not.
  ExecutionEngine *engine = static_cast<ExecutionEngine *>(e->llvm_engine);
  if (engine == nullptr) {
    engine = builder->create();
    if (engine == nullptr) {
      return 1;
    }
    e->llvm_engine = engine;
  } else {
    engine->addModule(std::move(pModule));
  }

  return 0;
}

void DeleteEngine(Engine *e) {
  // TODO: release llvm resource by call proper llvm dispose function.
  delete static_cast<MemoryManager *>(e->llvm_mem_manager);
  delete static_cast<EngineBuilder *>(e->llvm_builder);
  delete static_cast<legacy::PassManager *>(e->llvm_pass_manager);
  delete static_cast<LLVMContext *>(e->llvm_context);
  free(e);
}

int RunFunction(Engine *e, const char *funcName, size_t len,
                const uint8_t *data) {
  Module *module = static_cast<Module *>(e->llvm_main_module);
  ExecutionEngine *engine = static_cast<ExecutionEngine *>(e->llvm_engine);

  // finalize.
  engine->finalizeObject();
  engine->runStaticConstructorsDestructors(false);

  // run.
  Function *func = module->getFunction(StringRef(funcName));
  if (func == nullptr) {
    char msg[128];
    snprintf(msg, 128, "%s function not found.", funcName);
    LogError(msg);
    return -1;
  }

  (void)engine->getPointerToFunction(func);

  FunctionType *funcType = func->getFunctionType();
  // Run func.
  std::vector<GenericValue> args;

  GenericValue argLen;
  if (funcType->getNumParams() > 0) {
    argLen.IntVal = APInt(sizeof(size_t) * 8, len);
    args.push_back(argLen);

    GenericValue argData;
    argData.PointerVal = (PointerTy)data;
    args.push_back(argData);
  }

  auto ret = engine->runFunction(func, args);

  if (funcType->getReturnType()->isIntegerTy()) {
    return (int)(ret.IntVal.getSExtValue());
  } else {
    return 0;
  }
}

void AddNamedFunction(Engine *e, const char *funcName, void *address) {
  MemoryManager *mm = static_cast<MemoryManager *>(e->llvm_mem_manager);
  mm->addNamedFunction(funcName, address);
}
