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

#include <stdio.h>
#include <vector>

#include <llvm/ExecutionEngine/ExecutionEngine.h>
#include <llvm/ExecutionEngine/GenericValue.h>
#include <llvm/ExecutionEngine/SectionMemoryManager.h>
#include <llvm/IR/LLVMContext.h>
#include <llvm/IR/LegacyPassManager.h>
#include <llvm/IR/Module.h>
#include <llvm/IRReader/IRReader.h>
#include <llvm/Support/SourceMgr.h>
#include <llvm/Support/TargetSelect.h>
#include <llvm/Support/raw_ostream.h>
#include <llvm/Transforms/Scalar.h>
#include <llvm/Transforms/Scalar/GVN.h>

using namespace llvm;

void help(char *name) {
  printf("%s [IR file] [Func Name] [Args...]\n", name);
  exit(1);
}

extern "C" int roll_dice() { return 123; }

static void reportError(SMDiagnostic &Err, const char *ProgName) {
  Err.print(ProgName, errs());
  exit(1);
}

int main(int argc, char *argv[]) {
  if (argc < 3) {
    help(argv[0]);
  }

  char *progName = argv[0];
  char *filePath = argv[1];
  char *funcName = argv[2];

  // Initialization.
  InitializeNativeTarget();
  InitializeNativeTargetAsmParser();
  InitializeNativeTargetAsmPrinter();
  LLVMLinkInMCJIT();

  // Parse IR.
  LLVMContext context;
  SMDiagnostic err;

  std::unique_ptr<Module> pModule =
      parseIRFile(StringRef(filePath), err, context);
  Module *module = pModule.get();
  if (module == nullptr) {
    reportError(err, progName);
  }

  // Create PassManager.
  legacy::PassManager *passMgr = new legacy::PassManager();
  passMgr->add(createConstantPropagationPass());
  passMgr->add(createInstructionCombiningPass());
  passMgr->add(createPromoteMemoryToRegisterPass());
  passMgr->add(createCFGSimplificationPass());
  passMgr->add(createDeadCodeEliminationPass());
  passMgr->add(createGVNPass());
  passMgr->run(*module);

  // Create EngineBuilder.
  std::string errMsg;

  EngineBuilder builder(std::move(pModule));
  builder.setErrorStr(&errMsg);
  builder.setEngineKind(EngineKind::JIT);
  builder.setUseOrcMCJITReplacement(false);

  // Enable MCJIT.
  SectionMemoryManager *rtDyldMM = new SectionMemoryManager();
  builder.setMCJITMemoryManager(std::unique_ptr<RTDyldMemoryManager>(rtDyldMM));

  builder.setOptLevel(CodeGenOpt::Default);

  // Create ExecutionEngine.
  std::unique_ptr<ExecutionEngine> engine(builder.create());
  if (!engine) {
    if (!errMsg.empty()) {
      errs() << argv[0] << ": error creating EE: " << errMsg << "\n";
    } else {
      errs() << argv[0] << ": unknown error creating EE!\n";
    }
    exit(1);
  }

  Function *func = module->getFunction(StringRef(funcName));
  if (func == nullptr) {
    errs() << '\'' << funcName << "\' function not found in module.\n";
    return -1;
  }

  // Run static contructors.
  engine->finalizeObject();
  engine->runStaticConstructorsDestructors(false);
  (void)engine->getPointerToFunction(func);

  // Clear instruction cache.
  rtDyldMM->invalidateInstructionCache();

  // Run func.
  ArrayRef<GenericValue> args;
  engine->runFunction(func, args);

  // Run static destructors.
  engine->runStaticConstructorsDestructors(true);

  return 0;
}
