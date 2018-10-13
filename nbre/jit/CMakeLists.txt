
set(nbre_src ${nbre_src} OrcLazyJIT.cpp jit_driver.cpp PARENT_SCOPE
  )
add_library(nbre_jit OrcLazyJIT.cpp jit_driver.cpp
  )

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
target_link_libraries(nbre_jit ${llvm_libs})

enable_clang_tidy(nbre_jit)
