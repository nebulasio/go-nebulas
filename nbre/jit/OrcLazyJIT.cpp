//===- OrcLazyJIT.cpp - Basic Orc-based JIT for lazy execution ------------===//
//
//                     The LLVM Compiler Infrastructure
//
// This file is distributed under the University of Illinois Open Source
// License. See LICENSE.TXT for details.
//
//===----------------------------------------------------------------------===//

#include "OrcLazyJIT.h"
#include "common/common.h"
#include "jit/jit_exception.h"
#include "llvm/ADT/Triple.h"
#include "llvm/ExecutionEngine/ExecutionEngine.h"
#include "llvm/Support/CodeGen.h"
#include "llvm/Support/CommandLine.h"
#include "llvm/Support/DynamicLibrary.h"
#include "llvm/Support/ErrorHandling.h"
#include "llvm/Support/FileSystem.h"
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <iostream>
#include <system_error>

namespace {

enum class DumpKind {
  NoDump,
  DumpFuncsToStdOut,
  DumpModsToStdOut,
  DumpModsToDisk
};

} // end anonymous namespace

llvm::OrcLazyJIT::TransformFtor llvm::OrcLazyJIT::createDebugDumper() {
  DumpKind OrcDumpKind = DumpKind::NoDump;
  switch (OrcDumpKind) {
  case DumpKind::NoDump:
    return [](std::shared_ptr<Module> M) { return M; };

  case DumpKind::DumpFuncsToStdOut:
    return [](std::shared_ptr<Module> M) {
      std::cout << "[ " << std::endl;

      for (const auto &F : *M) {
        if (F.isDeclaration()) {
          continue;
        }

        if (F.hasName()) {
          std::string Name(F.getName());
          std::cout << Name << std::endl;
        } else {
          std::cout << "<anon> " << std::endl;
        }
      }

      std::cout << "]\n" << std::endl;
      return M;
    };

  case DumpKind::DumpModsToStdOut:
    return [](std::shared_ptr<Module> M) {
      outs() << "----- Module Start -----\n"
             << *M << "----- Module End -----\n";

      return M;
    };

  case DumpKind::DumpModsToDisk:
    return [](std::shared_ptr<Module> M) {
      std::error_code EC;
      raw_fd_ostream Out(M->getModuleIdentifier() + ".ll", EC, sys::fs::F_Text);
      if (EC) {
        errs() << "Couldn't open " << M->getModuleIdentifier()
               << " for dumping.\nError:" << EC.message() << "\n";
        exit(1);
      }
      Out << *M;
      return M;
    };
  }
  llvm_unreachable("Unknown DumpKind");
}

// Defined in lli.cpp.
// CodeGenOpt::Level getOptLevel();

template <typename PtrTy>
static PtrTy fromTargetAddress(llvm::JITTargetAddress Addr) {
  return reinterpret_cast<PtrTy>(static_cast<uintptr_t>(Addr));
}

int llvm::runOrcLazyJIT(neb::core::driver *d,
                        std::vector<std::unique_ptr<Module>> Ms,
                        const std::string &func_name, void *param) {
  // Grab a target machine and try to build a factory function for the
  // target-specific Orc callback manager.
  EngineBuilder EB;
  EB.setOptLevel(CodeGenOpt::Default);
  auto TM = std::unique_ptr<TargetMachine>(EB.selectTarget());
  Triple T(TM->getTargetTriple());
  auto CompileCallbackMgr = orc::createLocalCompileCallbackManager(T, 0);

  // If we couldn't build the factory function then there must not be a callback
  // manager for this target. Bail out.
  if (!CompileCallbackMgr) {
    LOG(ERROR) << "No callback manager available for target '"
               << TM->getTargetTriple().str() << "'.\n";
    throw neb::jit_internal_failure("No callback manager available for target");
  }

  auto IndirectStubsMgrBuilder = orc::createLocalIndirectStubsManagerBuilder(T);

  // If we couldn't build a stubs-manager-builder for this target then bail out.
  if (!IndirectStubsMgrBuilder) {
    LOG(ERROR) << "No indirect stubs manager available for target '"
               << TM->getTargetTriple().str() << "'.\n";
    throw neb::jit_internal_failure(
        "No indirect stubs manager available for target");
  }

  // Everything looks good. Build the JIT.
  bool OrcInlineStubs = true;
  OrcLazyJIT J(std::move(TM), std::move(CompileCallbackMgr),
               std::move(IndirectStubsMgrBuilder), OrcInlineStubs);

  // Add the module, look up main and run it.
  for (auto &M : Ms) {
    // outs() << *(M.get());
    outs().flush();

    cantFail(J.addModule(std::shared_ptr<Module>(std::move(M))), nullptr);
  }

  if (auto MainSym =
          J.findSymbol(std::string(func_name, std::allocator<char>()))) {
    using MainFnPtr = int (*)(neb::core::driver *, void *);
    auto Main =
        fromTargetAddress<MainFnPtr>(cantFail(MainSym.getAddress(), nullptr));
    return Main(d, param);
  } else if (auto Err = MainSym.takeError()) {
    logAllUnhandledErrors(std::move(Err), llvm::errs(), "");
    throw neb::jit_internal_failure("Unhandled errors");
  } else {
    LOG(ERROR) << "Could not find target function.\n";
    throw neb::jit_internal_failure("Could not find target function");
  }

  return 1;
}
