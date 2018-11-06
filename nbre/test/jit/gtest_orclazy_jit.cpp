#include "common/common.h"
#include "common/configuration.h"
#include "common/util/byte.h"
#include "fs/util.h"
#include "jit/OrcLazyJIT.h"
#include "jit/jit_driver.h"
#include "jit/jit_engine.h"
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

std::vector<std::unique_ptr<llvm::Module>>
get_modules(llvm::LLVMContext &context) {

  std::vector<std::unique_ptr<llvm::Module>> modules;
  llvm::SMDiagnostic err;
  std::string path = neb::fs::join_path(
      neb::configuration::instance().root_dir(), "test/data/test.bc");
  modules.push_back(llvm::parseIRFile(path, err, context));
  return modules;
}
TEST(test_jit, doule_addModule) {

  llvm::InitializeNativeTarget();
  llvm::InitializeNativeTargetAsmPrinter();
  llvm::sys::Process::PreventCoreFiles();
  std::string errMsg;
  llvm::sys::DynamicLibrary::LoadLibraryPermanently(nullptr, &errMsg);
  {
    llvm::LLVMContext context;
    auto modules = get_modules(context);
    neb::jit::jit_engine je;
    je.init(std::move(modules), "_Z9test_funcPN3neb4core6driverEPv");
    je.run<int, neb::core::driver *, void *>(nullptr, nullptr);
  }
  {
    llvm::LLVMContext context;
    auto modules = get_modules(context);
    neb::jit::jit_engine je;
    je.init(std::move(modules), "_Z9test_funcPN3neb4core6driverEPv");
    je.run<int, neb::core::driver *, void *>(nullptr, nullptr);
  }
  {
    llvm::LLVMContext context;
    auto modules = get_modules(context);
    neb::jit::jit_engine je;
    je.init(std::move(modules), "_Z9test_funcPN3neb4core6driverEPv");
    je.run<int, neb::core::driver *, void *>(nullptr, nullptr);
  }

  llvm::llvm_shutdown();
}
