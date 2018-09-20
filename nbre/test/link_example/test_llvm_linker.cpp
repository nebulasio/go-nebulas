#include "OrcLazyJIT.h"
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

#ifdef __CYGWIN__
#include <cygwin/version.h>
#if defined(CYGWIN_VERSION_DLL_MAJOR) && CYGWIN_VERSION_DLL_MAJOR < 1007
#define DO_NOTHING_ATEXIT 1
#endif
#endif

using namespace llvm;


LLVM_ATTRIBUTE_NORETURN
static void reportError(SMDiagnostic Err, const char *ProgName) {
  Err.print(ProgName, errs());
  exit(1);
}

//===----------------------------------------------------------------------===//
// main Driver function
//
int main(int argc, char **argv, char *const *envp) {
  sys::PrintStackTraceOnErrorSignal(argv[0]);
  PrettyStackTraceProgram X(argc, argv);

  atexit(llvm_shutdown); // Call llvm_shutdown() on exit.

  // if (argc > 1)
  // ExitOnErr.setBanner(std::string(argv[0]) + ": ");

  // If we have a native target, initialize it to ensure it is linked in and
  // usable by the JIT.
  InitializeNativeTarget();
  InitializeNativeTargetAsmPrinter();
  // InitializeNativeTargetAsmParser();

  sys::Process::PreventCoreFiles();

  LLVMContext Context;

  // Load the bitcode...
  SMDiagnostic Err;
  std::string InputFile =
      "/home/xuepeng/git/go-nebulas/nbre/test/link_example/foo.bc";
  std::string sopath =
      "/home/xuepeng/git/go-nebulas/nbre/test/link_example/libtest_bar.so";
  llvm::sys::DynamicLibrary::LoadLibraryPermanently(sopath.c_str());
  std::unique_ptr<Module> Owner = parseIRFile(InputFile, Err, Context);
  Module *Mod = Owner.get();
  if (!Mod)
    reportError(Err, argv[0]);

  std::vector<std::unique_ptr<Module>> Ms;
  Ms.push_back(std::move(Owner));
  // for (auto &ExtraMod : ExtraModules) {
  // Ms.push_back(parseIRFile(ExtraMod, Err, Context));
  // if (!Ms.back())
  // reportError(Err, argv[0]);
  //}
  std::vector<std::string> Args;
  Args.push_back(InputFile);
  // for (auto &Arg : InputArgv)
  // Args.push_back(Arg);
  return runOrcLazyJIT(std::move(Ms), Args);
}
