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
#include <thread>

std::string gen_key(const std::vector<std::shared_ptr<nbre::NBREIR>> &irs,
                    const std::string &func_name) {
  std::stringstream ss;
  for (auto &m : irs) {
    ss << m->name() << m->version();
  }
  ss << func_name;
  return ss.str();
}

void run_ir_exit(const std::shared_ptr<nbre::NBREIR> &ir_ptr) {
  std::vector<std::shared_ptr<nbre::NBREIR>> irs;
  irs.push_back(ir_ptr);
  std::string key = gen_key(irs, "_Z9test_funcPN3neb4core6driverEPv");
  neb::jit_driver::instance().run<neb::core::driver *, void *>(
      key, irs, "_Z9test_funcPN3neb4core6driverEPv", nullptr);
  for (int i = 0; i < 1000; i++) {
    neb::jit_driver::instance().run_if_exists<int, neb::core::driver *, void *>(
        ir_ptr, "_Z9test_funcPN3neb4core6driverEPv", nullptr, nullptr);
  }
}

TEST(test_jit, irs_file) {
  std::ifstream ifs;
  ifs.open("../test/data/test.bc", std::ios::in | std::ios::binary);
  ifs.seekg(0, ifs.end);
  std::ifstream::pos_type size = ifs.tellg();

  neb::util::bytes buf(size);

  ifs.seekg(0, ifs.beg);
  ifs.read((char *)buf.value(), buf.size());

  nbre::NBREIR ir_info;
  ir_info.set_ir(neb::util::byte_to_string(buf));
  auto ir_ptr = std::make_shared<nbre::NBREIR>(ir_info);

  run_ir_exit(ir_ptr);

  std::thread t1(run_ir_exit, ir_ptr);
  std::thread t2(run_ir_exit, ir_ptr);

  t1.join();
  t2.join();
}

std::vector<std::unique_ptr<llvm::Module>>
get_modules(llvm::LLVMContext &context) {

  std::vector<std::unique_ptr<llvm::Module>> modules;
  llvm::SMDiagnostic err;
  std::string path = neb::fs::join_path(
      neb::configuration::instance().nbre_root_dir(), "test/data/test.bc");
  modules.push_back(llvm::parseIRFile(path, err, context));
  return modules;
}
void run_module() {
  for (int i = 0; i < 10; i++) {
    llvm::LLVMContext context;
    std::vector<std::unique_ptr<llvm::Module>> modules;
    llvm::SMDiagnostic err;
    std::string path = neb::fs::join_path(
        neb::configuration::instance().nbre_root_dir(), "test/data/test.bc");
    modules.push_back(llvm::parseIRFile(path, err, context));
    neb::jit::jit_engine je;
    je.init(std::move(modules), "_Z9test_funcPN3neb4core6driverEPv");
    je.run<int, neb::core::driver *, void *>(nullptr, nullptr);
  }
}

// TEST(test_jit, doule_addModule) {

// llvm::InitializeNativeTarget();
// llvm::InitializeNativeTargetAsmPrinter();
// llvm::sys::Process::PreventCoreFiles();
// std::string errMsg;
// llvm::sys::DynamicLibrary::LoadLibraryPermanently(nullptr, &errMsg);

// std::thread t_1(run_module);
// std::thread t_2(run_module);
// t_1.join();
// t_2.join();

//{
// llvm::LLVMContext context;
// auto modules = get_modules(context);
// neb::jit::jit_engine je;
// je.init(std::move(modules), "_Z9test_funcPN3neb4core6driverEPv");
// je.run<int, neb::core::driver *, void *>(nullptr, nullptr);
//}
//{
// llvm::LLVMContext context;
// auto modules = get_modules(context);
// neb::jit::jit_engine je;
// je.init(std::move(modules), "_Z9test_funcPN3neb4core6driverEPv");
// je.run<int, neb::core::driver *, void *>(nullptr, nullptr);
//}
//{
// llvm::LLVMContext context;
// auto modules = get_modules(context);
// neb::jit::jit_engine je;
// je.init(std::move(modules), "_Z9test_funcPN3neb4core6driverEPv");
// je.run<int, neb::core::driver *, void *>(nullptr, nullptr);
//}

// llvm::llvm_shutdown();
//}

void Run_One(const std::string &path, const std::string &func_name) {
  std::ifstream ifs;
  ifs.open(path.c_str(), std::ios::in | std::ios::binary);
  ifs.seekg(0, ifs.end);
  std::ifstream::pos_type size = ifs.tellg();

  neb::util::bytes buf(size);

  ifs.seekg(0, ifs.beg);
  ifs.read((char *)buf.value(), buf.size());

  nbre::NBREIR ir_info;
  ir_info.set_ir(neb::util::byte_to_string(buf));
  auto ir_ptr = std::make_shared<nbre::NBREIR>(ir_info);

  std::vector<std::shared_ptr<nbre::NBREIR>> irs;
  irs.push_back(ir_ptr);
  std::string key = gen_key(irs, func_name.c_str());
  std::cout << "Run_One: before run" << std::endl;
  neb::jit_driver::instance().run<neb::core::driver *, void *>(
      key, irs, func_name.c_str(), nullptr);
  neb::jit_driver::instance().run_if_exists<int, neb::core::driver *, void *>(
      ir_ptr, func_name.c_str(), nullptr, nullptr);
}

TEST(test_jit, error_functionName_irs_file) {
  Run_One("../test/data/test.bc", "_Z9test_funcPN3neb4core6driverEPv12");
}

TEST(test_jit, another_irs_file) {
  Run_One("../bin/jit_test_1.bc", "_Z10jit_test_1PN3neb4core6driverEPv");
}

void Run_One_1000(const std::string &path, const std::string &func_name) {
  for (int i = 0; i < 1000; ++i) {
    Run_One(path, func_name);
  }
}

TEST(test_jit, multi_thread) {
  std::thread t1(Run_One_1000, "../test/data/test.bc",
                 "_Z9test_funcPN3neb4core6driverEPv");
  std::thread t2(Run_One_1000, "../bin/jit_test_1.bc",
                 "_Z10jit_test_1PN3neb4core6driverEPv");
  std::thread t3(Run_One_1000, "../test/data/test.bc",
                 "_Z9test_funcPN3neb4core6driverEP1");
  std::thread t4(Run_One_1000, "../bin/jit_test_1.bc",
                 "_Z10jit_test_1PN3neb4core6driverEP1");
  t1.join();
  t2.join();
  t3.join();
  t4.join();
}
