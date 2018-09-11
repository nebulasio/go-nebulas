test_llvm_linker: test_llvm_linker.cpp
	clang++ -g -O3 -std=c++11 -fno-exceptions -fno-rtti test_llvm_linker.cpp `llvm-config --cxxflags --ldflags --libs --system-libs core executionengine interpreter mc support nativecodegen irreader linker` -lffi -o llvm_app_2 
