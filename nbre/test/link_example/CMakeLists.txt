
add_executable(link_example test_llvm_linker.cpp)

set_target_properties(
  link_example PROPERTIES
  CXX_STANDARD 11
  CXX_STANDARD_REQUIRED ON
  COMPILE_FLAGS "${WARNING_FLAGS}"
)

if(CLANG_TIDY_EXE)
  set_target_properties(
    link_example PROPERTIES
    CXX_CLANG_TIDY "${DO_CLANG_TIDY}"
)
endif()

llvm_map_components_to_libnames(llvm_libs core executionengine interpreter mc support nativecodegen irreader linker)
target_link_libraries(link_example ${llvm_libs})
