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

#include "common/common.h"
#include "common/configuration.h"
#include "common/util/byte.h"
#include "jit/OrcLazyJIT.h"
#include "jit/jit_driver.h"
#include "llvm/ADT/Triple.h"
#include "llvm/ExecutionEngine/ExecutionEngine.h"
#include "llvm/IR/LLVMContext.h"
#include "llvm/IR/Module.h"
#include "llvm/IRReader/IRReader.h"
#include "llvm/Support/CodeGen.h"
#include "llvm/Support/CommandLine.h"
#include "llvm/Support/DynamicLibrary.h"
#include "llvm/Support/ErrorHandling.h"
#include "llvm/Support/FileSystem.h"
#include "llvm/Support/MemoryBuffer.h"
#include "llvm/Support/SourceMgr.h"
#include "llvm/Support/TargetSelect.h"
#include "llvm/Support/raw_ostream.h"
#include <fstream>
#include <gtest/gtest.h>

TEST(test_jit, simple) { neb::jit_driver d; }

TEST(test_jit, doule_addModule) {

  llvm::InitializeNativeTarget();
  llvm::InitializeNativeTargetAsmPrinter();
  llvm::sys::Process::PreventCoreFiles();

  std::string errMsg;
  auto load =
      llvm::sys::DynamicLibrary::LoadLibraryPermanently(nullptr, &errMsg);
  EXPECT_TRUE(!load);

  std::ifstream ifs;
  ifs.open("./xx.bc", std::ios::in | std::ios::binary);
  ifs.seekg(0, ifs.end);
  std::ifstream::pos_type size = ifs.tellg();

  neb::util::bytes buf(size);

  ifs.seekg(0, ifs.beg);
  ifs.read((char *)buf.value(), buf.size());

  nbre::NBREIR ir_info;
  ir_info.set_ir(neb::util::byte_to_string(buf));
  auto ir_ptr = std::make_shared<nbre::NBREIR>(ir_info);
  std::vector<std::shared_ptr<nbre::NBREIR>> v({ir_ptr});

  std::vector<std::unique_ptr<llvm::Module>> modules;
  llvm::LLVMContext m_context;
  for (const auto &ir : v) {
    std::string ir_str = ir->ir();
    llvm::StringRef sr(ir_str);
    auto mem_buf = llvm::MemoryBuffer::getMemBuffer(sr, "", false);
    llvm::SMDiagnostic err;
    auto module =
        llvm::parseIR(mem_buf->getMemBufferRef(), err, m_context, true);
    std::cout << "err msg: " << err.getMessage().data() << std::endl;
    modules.push_back(
        llvm::parseIR(mem_buf->getMemBufferRef(), err, m_context, true));
  }
  std::cout << "end" << std::endl;
  llvm::runOrcLazyJIT(nullptr, std::move(modules), "", nullptr);

  // llvm::LLVMContext context;
  // llvm::SMDiagnostic error;
  // std::vector<std::unique_ptr<llvm::Module>> modules;
  // auto module = llvm::parseIRFile("/home/pluo/luopeng.bc", error, context);
  // std::cout << error.getMessage().data() << std::endl;
  // EXPECT_EQ(error.getMessage().data(), "");
  // EXPECT_TRUE(module);
  // modules.push_back(llvm::parseIRFile("/home/pluo/luopeng.bc", error,
  // context));
  // modules.push_back(
  //    llvm::parseIR(mem_buf->getMemBufferRef(), error, context, true));

  // llvm::EngineBuilder EB;
  // EB.setOptLevel(llvm::CodeGenOpt::Default);
  // auto TM = std::unique_ptr<llvm::TargetMachine>(EB.selectTarget());
  // llvm::Triple T(TM->getTargetTriple());
  // auto CompileCallbackMgr = llvm::orc::createLocalCompileCallbackManager(T,
  // 0); EXPECT_TRUE(CompileCallbackMgr);

  // auto IndirectStubsMgrBuilder =
  // llvm::orc::createLocalIndirectStubsManagerBuilder(T);
  // EXPECT_TRUE(IndirectStubsMgrBuilder);

  //// Everything looks good. Build the JIT.
  // bool OrcInlineStubs = true;
  // llvm::OrcLazyJIT J(std::move(TM), std::move(CompileCallbackMgr),
  // std::move(IndirectStubsMgrBuilder), OrcInlineStubs);

  //// Add the module, look up main and run it.
  // for (auto &M : modules) {
  //// outs() << *(M.get());
  // llvm::outs().flush();

  // cantFail(J.addModule(std::shared_ptr<llvm::Module>(std::move(M))),
  // nullptr);
  //}
  std::cout << "end" << std::endl;
}
