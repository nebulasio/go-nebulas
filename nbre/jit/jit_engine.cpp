// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify
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
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//
#include "jit/jit_engine.h"

namespace neb {
namespace jit {
using namespace llvm;

void jit_engine::init(std::vector<std::unique_ptr<Module>> ms,
                      const std::string &func_name) {
  m_modules = std::move(ms);
  m_func_name = func_name;
  // Grab a target machine and try to build a factory function for the
  // target-specific Orc callback manager.
  m_EB = std::make_unique<llvm::EngineBuilder>();
  m_EB->setOptLevel(CodeGenOpt::Default);
  auto TM = std::unique_ptr<TargetMachine>(m_EB->selectTarget());
  m_T = std::make_unique<Triple>(TM->getTargetTriple());

  auto CompileCallbackMgr = orc::createLocalCompileCallbackManager(*m_T, 0);
  // If we couldn't build the factory function then there must not be a
  // callback manager for this target. Bail out.
  if (!CompileCallbackMgr) {
    LOG(ERROR) << "No callback manager available for target '"
               << TM->getTargetTriple().str() << "'.\n";
    throw neb::jit_internal_failure("No callback manager available for target");
  }

  auto IndirectStubsMgrBuilder =
      orc::createLocalIndirectStubsManagerBuilder(*m_T);

  // If we couldn't build a stubs-manager-builder for this target then bail
  // out.
  if (!IndirectStubsMgrBuilder) {
    LOG(ERROR) << "No indirect stubs manager available for target '"
               << TM->getTargetTriple().str() << "'.\n";
    throw neb::jit_internal_failure(
        "No indirect stubs manager available for target");
  }

  // Everything looks good. Build the JIT.
  bool OrcInlineStubs = true;
  m_jit = std::make_unique<OrcLazyJIT>(
      std::move(TM), std::move(CompileCallbackMgr),
      std::move(IndirectStubsMgrBuilder), OrcInlineStubs);

  // Add the module, look up main and run it.
  for (auto &M : m_modules) {
    outs().flush();

    try {
      cantFail(m_jit->addModule(std::shared_ptr<Module>(std::move(M))),
               nullptr);
    } catch (std::exception &e) {
      LOG(ERROR) << e.what();
    }
  }
  try {
    m_main_sym = std::make_unique<llvm::JITSymbol>(
        m_jit->findSymbol(std::string(m_func_name, std::allocator<char>())));
    if (*m_main_sym) {
      return;
    } else if (auto Err = m_main_sym->takeError()) {
      logAllUnhandledErrors(std::move(Err), llvm::errs(), "");
      throw neb::jit_internal_failure("Unhandled errors");
    } else {
      LOG(ERROR) << "Could not find target function.\n";
      throw neb::jit_internal_failure("Could not find target function");
    }
  } catch (std::exception &e) {
    LOG(INFO) << e.what();
    }
}
} // namespace jit
} // namespace neb
