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

jit_driver::jit_driver() {
  // llvm::sys::PrintStackTraceOnErrorSignal(
  // configuration::instance().exec_name(), false);
  llvm::InitializeNativeTarget();
  llvm::InitializeNativeTargetAsmPrinter();
  llvm::sys::Process::PreventCoreFiles();
  std::string errMsg;
  llvm::sys::DynamicLibrary::LoadLibraryPermanently(nullptr, &errMsg);
}

jit_driver::~jit_driver() { llvm::llvm_shutdown(); }

bool jit_driver::find_mangling(llvm::Module *M, const std::string &func_name,
                               std::string &mangling_name) {
  auto contains = [](const std::string &str,
                     const std::string &substr) -> bool {
    size_t str_len = str.size();
    size_t substr_len = substr.size();

    size_t j = 0;
    for (size_t i = 0; i < str_len && j < substr_len; i++) {
      if (str[i] == substr[j]) {
        j++;
      }
    }
    return j == substr_len;
  };

  for (auto &func : M->functions()) {
    std::string name = func.getName().data();
    if (contains(name, func_name)) {
      mangling_name = name;
      return true;
    }
  }
  return false;
}

std::unique_ptr<jit_driver::jit_context>
jit_driver::make_context(const std::vector<nbre::NBREIR> &irs,
                         const std::string &func_name) {
  std::unique_ptr<jit_context> ret = std::make_unique<jit_context>();

  std::string mangling_name;
  std::vector<std::unique_ptr<llvm::Module>> modules;
  for (const auto &ir : irs) {
    std::string ir_str = ir.ir();
    llvm::StringRef sr(ir_str);
    auto mem_buf = llvm::MemoryBuffer::getMemBuffer(sr, "", false);
    llvm::SMDiagnostic err;

    auto module =
        llvm::parseIR(mem_buf->getMemBufferRef(), err, ret->m_context, true);
    find_mangling(module.get(), func_name, mangling_name);
    if (nullptr == module) {
      LOG(ERROR) << "Module broken";
    } else {
      modules.push_back(std::move(module));
    }
  }
  ret->m_jit.init(std::move(modules), mangling_name);
  return std::move(ret);
}

void jit_driver::timer_callback() {
  std::unique_lock<std::mutex> _l(m_mutex);
  std::vector<std::string> keys;
  for (auto &it : m_jit_instances) {
    it.second->m_time_counter--;
    if (it.second->m_time_counter < 0) {
      keys.push_back(it.first);
    }
  }
  for (auto &s : keys) {
    m_jit_instances.erase(s);
  }
}

std::string jit_driver::gen_key(const std::vector<nbre::NBREIR> &irs,
                                const std::string &func_name) {
  std::stringstream ss;
  for (auto &m : irs) {
    ss << m.name() << m.version();
  }
  ss << func_name;
  return ss.str();
}

void jit_driver::shrink_instances() {
  if (m_jit_instances.size() < 16)
    return;

  int32_t min_count = 30 * 60 + 1;
  std::string min_key;
  for (auto &pair : m_jit_instances) {
    if (pair.second->m_time_counter < min_count && !pair.second->m_using) {
      min_count = pair.second->m_time_counter;
      min_key = pair.first;
    }
  }
  if (!min_key.empty()) {
    m_jit_instances.erase(min_key);
  }
}
} // namespace neb
