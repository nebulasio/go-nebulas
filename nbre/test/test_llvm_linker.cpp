#include "llvm/ADT/STLExtras.h"
#include "llvm/ExecutionEngine/ExecutionEngine.h"
#include "llvm/ExecutionEngine/GenericValue.h"
#include "llvm/ExecutionEngine/Interpreter.h"
#include "llvm/IR/Argument.h"
#include "llvm/IR/BasicBlock.h"
#include "llvm/IR/Constants.h"
#include "llvm/IR/DerivedTypes.h"
#include "llvm/IR/Function.h"
#include "llvm/IR/IRBuilder.h"
#include "llvm/IR/Instructions.h"
#include "llvm/IR/LLVMContext.h"
#include "llvm/IR/Module.h"
#include "llvm/IR/Type.h"
#include "llvm/Support/Casting.h"
#include "llvm/Support/ManagedStatic.h"
#include "llvm/Support/TargetSelect.h"
#include "llvm/Support/raw_ostream.h"
#include "llvm/Support/SourceMgr.h"
#include "llvm/IRReader/IRReader.h"
#include "llvm/Linker/Linker.h"

#include <algorithm>
#include <cassert>
#include <memory>
#include <vector>

using namespace llvm;

typedef std::unique_ptr<Module> LLVMModule;
typedef std::unique_ptr<ExecutionEngine> LLVMEngine;

LLVMContext _context;
SMDiagnostic _error;


void initialize() {
    InitializeNativeTarget();
}


void printIR(LLVMModule& module) {
    if (module != NULL) {
        outs() << *(module.get());    
        outs().flush();
    }
}


LLVMModule getModuleFromIRFile(char* fileName) {
    return parseIRFile(StringRef(fileName), _error, _context);
    //return getLazyIRFileModule(StringRef(fileName), _error, _context);
}


LLVMEngine createEngine(LLVMModule& module) {
    std::string error;

    ExecutionEngine* engine = EngineBuilder(std::move(module))
        .setErrorStr(&error)
        .create();

    return std::unique_ptr<ExecutionEngine> (engine); 
}


GenericValue runFunction(ExecutionEngine* engine, Function* funcMain) {
    GenericValue gv;
    if (NULL != funcMain) {
        std::vector<GenericValue> noargs;
        gv = engine->runFunction(funcMain, noargs);
    }

    return gv;
}


void printResult(const GenericValue& gv) {
    outs() << "Result: " << gv.IntVal << "\n";
}


void shutdownLLVM() {
    llvm_shutdown();
}


void linkModule(Linker& linker, LLVMModule& module) {
    bool error = linker.linkInModule(std::move(module), Linker::Flags::None & Linker::Flags::OverrideFromSrc);
    assert(error != true);
}


LLVMModule linkFooAndBarModule(LLVMModule& fooModule, LLVMModule& barModule) {
    LLVMModule compositeModule = make_unique<Module>("llvm-link", _context);
    Linker linker(*compositeModule);

    linkModule(linker, fooModule);
    linkModule(linker, barModule);
    
    return compositeModule;
}


int main(int argc, char * argv[]) {

    if (argc < 3) {
        errs() << "Expected an argument - IR file name\n";
        exit(1);
    }

    initialize();

    LLVMModule fooModule = getModuleFromIRFile(argv[1]);
    LLVMModule barModule = getModuleFromIRFile(argv[2]);

    LLVMModule compositeModule = linkFooAndBarModule(fooModule, barModule);

    printIR(compositeModule);

    Function* funcFoo = compositeModule->getFunction("main");

    LLVMEngine engine = createEngine(compositeModule);

    GenericValue gv = runFunction(engine.get(), funcFoo);

    printResult(gv);

    shutdownLLVM();

    return 0;
}


