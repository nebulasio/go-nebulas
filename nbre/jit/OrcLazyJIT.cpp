//===- OrcLazyJIT.cpp - Basic Orc-based JIT for lazy execution ------------===//
//
//                     The LLVM Compiler Infrastructure
//
// This file is distributed under the University of Illinois Open Source
// License. See LICENSE.TXT for details.
//
//===----------------------------------------------------------------------===//

#include "OrcLazyJIT.h"
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

llvm::Error llvm::OrcLazyJIT::addModule(std::shared_ptr<Module> M) {
  if (M->getDataLayout().isDefault())
    M->setDataLayout(DL);

  // Rename, bump linkage and record static constructors and destructors.
  // We have to do this before we hand over ownership of the module to the
  // JIT.
  std::vector<std::string> CtorNames, DtorNames;
  {
    unsigned CtorId = 0, DtorId = 0;
    for (auto Ctor : orc::getConstructors(*M)) {
      std::string NewCtorName = ("$static_ctor." + Twine(CtorId++)).str();
      Ctor.Func->setName(NewCtorName);
      Ctor.Func->setLinkage(GlobalValue::ExternalLinkage);
      Ctor.Func->setVisibility(GlobalValue::HiddenVisibility);
      CtorNames.push_back(mangle(NewCtorName));
    }
    for (auto Dtor : orc::getDestructors(*M)) {
      std::string NewDtorName = ("$static_dtor." + Twine(DtorId++)).str();
      Dtor.Func->setLinkage(GlobalValue::ExternalLinkage);
      Dtor.Func->setVisibility(GlobalValue::HiddenVisibility);
      DtorNames.push_back(mangle(Dtor.Func->getName()));
      Dtor.Func->setName(NewDtorName);
    }
  }

  // Symbol resolution order:
  //   1) Search the JIT symbols.
  //   2) Check for C++ runtime overrides.
  //   3) Search the host process (LLI)'s symbol table.
  if (!ModulesHandle) {
    auto Resolver = orc::createLambdaResolver(
        [this](const std::string &Name) -> JITSymbol {
          if (auto Sym = CODLayer.findSymbol(Name, true))
            return Sym;
          return CXXRuntimeOverrides.searchOverrides(Name);
        },
        [this](const std::string &Name) {
          if (auto Addr = RTDyldMemoryManager::getSymbolAddressInProcess(Name))
            return JITSymbol(Addr, JITSymbolFlags::Exported);
          return JITSymbol(nullptr);
        });

    // Add the module to the JIT.
    if (auto ModulesHandleOrErr =
            CODLayer.addModule(std::move(M), std::move(Resolver)))
      ModulesHandle = std::move(*ModulesHandleOrErr);
    else
      return ModulesHandleOrErr.takeError();

  } else if (auto Err = CODLayer.addExtraModule(*ModulesHandle, std::move(M)))
    return Err;

  // Run the static constructors, and save the static destructor runner for
  // execution when the JIT is torn down.
  orc::CtorDtorRunner<CODLayerT> CtorRunner(std::move(CtorNames),
                                            *ModulesHandle);
  if (auto Err = CtorRunner.runViaLayer(CODLayer))
    return Err;

  IRStaticDestructorRunners.emplace_back(std::move(DtorNames), *ModulesHandle);

  return Error::success();
}

// Defined in lli.cpp.
// CodeGenOpt::Level getOptLevel();


