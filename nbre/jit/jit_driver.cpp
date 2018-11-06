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

#include "jit/jit_driver.h"
#include "common/configuration.h"
#include "jit/OrcLazyJIT.h"

#include "jit/jit_exception.h"
#include "llvm/ADT/StringExtras.h"
#include "llvm/ADT/Triple.h"
#include "llvm/Bitcode/BitcodeReader.h"
#include "llvm/CodeGen/CommandFlags.def"
#include "llvm/CodeGen/LinkAllCodegenComponents.h"
#include "llvm/ExecutionEngine/GenericValue.h"
#include "llvm/ExecutionEngine/Interpreter.h"
#include "llvm/ExecutionEngine/JITEventListener.h"
#include "llvm/ExecutionEngine/MCJIT.h"
#include "llvm/ExecutionEngine/ObjectCache.h"
#include "llvm/ExecutionEngine/Orc/OrcRemoteTargetClient.h"
#include "llvm/ExecutionEngine/OrcMCJITReplacement.h"
#include "llvm/ExecutionEngine/SectionMemoryManager.h"
#include "llvm/IR/IRBuilder.h"
#include "llvm/IR/LLVMContext.h"
#include "llvm/IR/Module.h"
#include "llvm/IR/Type.h"
#include "llvm/IR/TypeBuilder.h"
#include "llvm/IRReader/IRReader.h"
#include "llvm/Object/Archive.h"
#include "llvm/Object/ObjectFile.h"
#include "llvm/Support/CommandLine.h"
#include "llvm/Support/Debug.h"
#include "llvm/Support/DynamicLibrary.h"
#include "llvm/Support/Format.h"
#include "llvm/Support/ManagedStatic.h"
#include "llvm/Support/MathExtras.h"
#include "llvm/Support/Memory.h"
#include "llvm/Support/MemoryBuffer.h"
#include "llvm/Support/Path.h"
#include "llvm/Support/PluginLoader.h"
#include "llvm/Support/PrettyStackTrace.h"
#include "llvm/Support/Process.h"
#include "llvm/Support/Program.h"
#include "llvm/Support/Signals.h"
#include "llvm/Support/SourceMgr.h"
#include "llvm/Support/TargetSelect.h"
#include "llvm/Support/raw_ostream.h"
#include "llvm/Transforms/Instrumentation.h"
#include <cerrno>

namespace neb {
namespace internal {

class jit_driver_impl {
public:
  jit_driver_impl() {
    // llvm::sys::PrintStackTraceOnErrorSignal(
    // configuration::instance().exec_name(), false);
    llvm::InitializeNativeTarget();
    llvm::InitializeNativeTargetAsmPrinter();
    llvm::sys::Process::PreventCoreFiles();
  }
  jit_driver_impl(const jit_driver_impl &) = delete;
  jit_driver_impl &operator=(const jit_driver_impl &) = delete;
  jit_driver_impl(jit_driver_impl &&) = delete;
  jit_driver_impl &&operator=(const jit_driver_impl &&) = delete;

  void run(core::driver *d,
           const std::vector<std::shared_ptr<nbre::NBREIR>> &irs,
           const std::string &func_name, void *param) {
    std::string errMsg;
    if (llvm::sys::DynamicLibrary::LoadLibraryPermanently(nullptr, &errMsg)) {
      LOG(ERROR) << errMsg;
      throw jit_internal_failure("failed to load local program");
    }
    std::vector<std::unique_ptr<llvm::Module>> modules;
    for (const auto &ir : irs) {
      std::string ir_str = ir->ir();
      llvm::StringRef sr(ir_str);
      auto mem_buf = llvm::MemoryBuffer::getMemBuffer(sr, "", false);
      llvm::SMDiagnostic err;

      modules.push_back(
          llvm::parseIR(mem_buf->getMemBufferRef(), err, m_context, true));
    }
    LOG(INFO) << " call llvm::runOrcLazyJIT";
    auto ret = llvm::runOrcLazyJIT(d, std::move(modules), func_name, param);
    LOG(INFO) << "jit return : " << ret;
  }

  void auto_run(const nbre::NBREIR &ir, const std::string &func_name,
                auth_table_t &auth_table) {

    std::string errMsg;
    if (llvm::sys::DynamicLibrary::LoadLibraryPermanently(nullptr, &errMsg)) {
      LOG(ERROR) << errMsg;
      throw jit_internal_failure("failed to load local program");
    }
    std::string ir_str = ir.ir();
    llvm::StringRef sr(ir_str);
    auto mem_buf = llvm::MemoryBuffer::getMemBuffer(sr, "", false);
    llvm::SMDiagnostic err;

    std::unique_ptr<llvm::Module> module =
        llvm::parseIR(mem_buf->getMemBufferRef(), err, m_context, true);
    LOG(INFO) << " call llvm::auto_runOrcLazyJIT";
    auth_table = llvm::auto_runOrcLazyJIT(std::move(module), func_name);
  }

  virtual ~jit_driver_impl() { llvm::llvm_shutdown(); }

protected:
  llvm::LLVMContext m_context;

}; // end class jit_driver_impl
} // end namespace internal

jit_driver::jit_driver() {
  m_impl = std::unique_ptr<internal::jit_driver_impl>(
      new internal::jit_driver_impl());
}

jit_driver::~jit_driver() = default;

void jit_driver::run(core::driver *d,
                     const std::vector<std::shared_ptr<nbre::NBREIR>> &irs,
                     const std::string &func_name, void *param) {
  m_impl->run(d, irs, func_name, param);
}

void jit_driver::auth_run(const nbre::NBREIR &ir, const std::string &func_name,
                          auth_table_t &auth_table) {
  m_impl->auto_run(ir, func_name, auth_table);
}
} // namespace neb
