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
#include "jit/jit_mangled_entry_point.h"
#include "jit/cpp_ir.h"
#include "llvm/IR/IRBuilder.h"
#include "llvm/IR/LLVMContext.h"
#include "llvm/IR/Module.h"
#include "llvm/IRReader/IRReader.h"
#include "llvm/Support/Memory.h"
#include "llvm/Support/MemoryBuffer.h"
//#include "llvm/Support/PluginLoader.h"
#include "common/configuration.h"
#include "llvm/Support/SourceMgr.h"
#include "llvm/Support/TargetSelect.h"
#include "llvm/Support/raw_ostream.h"

namespace neb {
namespace jit {
jit_mangled_entry_point::jit_mangled_entry_point()
    : m_mutex(), m_prog_slice(), m_mangled_entry_names() {
  init_prog_slice();
}

jit_mangled_entry_point::~jit_mangled_entry_point() {}

std::string
jit_mangled_entry_point::get_mangled_entry_name(const std::string &entry_name) {
  std::unique_lock<std::mutex> _l(m_mutex);
  if (m_mangled_entry_names.find(entry_name) != m_mangled_entry_names.end()) {
    return m_mangled_entry_names[entry_name];
  }
  gen_mangle_name_for_entry(entry_name);
  if (m_mangled_entry_names.find(entry_name) != m_mangled_entry_names.end()) {
    return m_mangled_entry_names[entry_name];
  } else {
    return std::string();
  }
}

void jit_mangled_entry_point::init_prog_slice() {
  m_prog_slice.insert(std::make_pair(
      configuration::instance().nr_func_name(),
      "#include \"runtime/nr/impl/nr_impl.h\"\n"
      "neb::rt::nr::nr_ret_type "
      "entry_point_nr(neb::compatible_uint64_t, "
      "neb::compatible_uint64_t){return neb::rt::nr::nr_ret_type(); }"));
  m_prog_slice.insert(
      std::make_pair(configuration::instance().dip_func_name(),
                     "#include \"runtime/dip/dip_impl.h\"\n"
                     "neb::rt::dip::dip_ret_type "
                     "entry_point_dip(neb::compatible_uint64_t)"
                     "{return neb::rt::dip::dip_ret_type(); }"));
  m_prog_slice.insert(
      std::make_pair(configuration::instance().auth_func_name(),
                     "#include <string>\n"
                     "#include <tuple>\n"
                     "#include <vector>\n"
                     "typedef std::tuple<std::string, std::string, uint64_t, "
                     "uint64_t> row_t;\n"
                     "std::vector<row_t> entry_point_auth(){return "
                     "std::vector<row_t>();}"));
}

void jit_mangled_entry_point::gen_mangle_name_for_entry(
    const std::string &entry_name) {
  if (m_prog_slice.find(entry_name) == m_prog_slice.end())
    return;

  auto auth_check_func = [](llvm::LLVMContext &context,
                            const llvm::Function &func) -> bool {
    if (func.isIntrinsic())
      return false;
    llvm::FunctionType *ft = func.getFunctionType();
    if (ft->getNumParams() > 1)
      return false;

    if (func.getReturnType()->isIntegerTy()) {
      return false;
    }
    return true;
  };

  auto nr_check_func = [](llvm::LLVMContext &context,
                          const llvm::Function &func) -> bool {
    if (func.isIntrinsic())
      return false;
    llvm::FunctionType *ft = func.getFunctionType();
    if (ft->getNumParams() < 2)
      return false;


    std::vector<llvm::Type *> ts;
    for (auto it = ft->param_begin(); it != ft->param_end(); it++) {
      ts.push_back(*it);
    }
    if (ts[ts.size() - 1] == ts[ts.size() - 2] &&
        ts[ts.size() - 1] == llvm::Type::getInt64Ty(context)) {
      return true;
    }
    return false;
  };
  auto dip_check_func = [](llvm::LLVMContext &context,
                           const llvm::Function &func) -> bool {
    if (func.isIntrinsic())
      return false;
    llvm::FunctionType *ft = func.getFunctionType();
    if (ft->getNumParams() < 1)
      return false;
    if (ft->getNumParams() >= 3)
      return false;

    std::vector<llvm::Type *> ts;
    for (auto it = ft->param_begin(); it != ft->param_end(); it++) {
      ts.push_back(*it);
    }
    if (func.getReturnType() == llvm::Type::getVoidTy(context) &&
        ts.size() != 2)
      return false;

    if (func.getReturnType()->isIntegerTy()) {
      return false;
    }

    if (ts[ts.size() - 1] == llvm::Type::getInt64Ty(context)) {
      return true;
    }
    return false;
  };

  std::function<bool(llvm::LLVMContext &, const llvm::Function &)> check_func;
  if (entry_name == configuration::instance().auth_func_name()) {
    check_func = auth_check_func;
  } else if (entry_name == configuration::instance().nr_func_name()) {
    check_func = nr_check_func;
  } else if (entry_name == configuration::instance().dip_func_name()) {
    check_func = dip_check_func;
  } else {
    return;
  }

  ::llvm::LLVMContext context;
  cpp::cpp_ir ir_gen(std::make_pair(entry_name, m_prog_slice[entry_name]));
  neb::bytes ir = ir_gen.llvm_ir_content();
  std::string ir_str = byte_to_string(ir);
  llvm::StringRef sr(ir_str);
  auto mem_buf = llvm::MemoryBuffer::getMemBuffer(sr, "", false);
  llvm::SMDiagnostic err;
  auto module = llvm::parseIR(mem_buf->getMemBufferRef(), err, context, true);
  if (nullptr == module) {
    LOG(ERROR) << "Module broken ";
  } else {
    for (auto &func : module->functions()) {
      if (check_func(context, func)) {
        std::string name = func.getName().data();
        m_mangled_entry_names[entry_name] = name;
      }
    }
  }
}

} // namespace jit
} // namespace neb
