
add_definitions(-fno-rtti)

add_executable(link_example test_llvm_linker.cpp OrcLazyJIT.cpp)
add_library(test_bar SHARED bar.cpp)

llvm_map_components_to_libnames(llvm_libs
  core
  executionengine
  interpreter
  mc mcjit
  support
  nativecodegen
  irreader
  linker
  orcjit
  runtimedyld
  target
  support
  option)
target_link_libraries(link_example ${llvm_libs})

enable_clang_tidy(nbre_fs)
