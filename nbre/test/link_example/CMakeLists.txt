
add_executable(link_example test_llvm_linker.cpp)

llvm_map_components_to_libnames(llvm_libs core executionengine interpreter mc support nativecodegen irreader linker)
target_link_libraries(link_example ${llvm_libs})

enable_clang_tidy(nbre_fs)
