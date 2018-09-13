cmake_minimum_required(VERSION 3.12)

project(nebulas-blockchain-runtime-environment)
find_program(
  CLANG_TIDY_EXE
  NAMES "clang-tidy"
  DOC "Path to clang-tidy executable"
)

if(NOT CLANG_TIDY_EXE)
  message(STATUS "clang-tidy not found.")
else()
  message(STATUS "clang-tidy found: ${CLANG_TIDY_EXE}")
  set(DO_CLANG_TIDY "${CLANG_TIDY_EXE}" "-checks=*,-clang-analyzer-alpha.*")
endif()

find_package(LLVM REQUIRED CONFIG)

include_directories(${LLVM_INCLUDE_DIRS})
add_definitions(${LLVM_DEFINITIONS})

function(enable_clang_tidy target_name)
  set_target_properties(
    ${target_name} PROPERTIES
    CXX_STANDARD 11
    CXX_STANDARD_REQUIRED ON
    COMPILE_FLAGS "${WARNING_FLAGS}"
  )
  if(CLANG_TIDY_EXE)
    set_target_properties(
      ${target_name} PROPERTIES
      CXX_CLANG_TIDY "${DO_CLANG_TIDY}"
  )
  endif()
endfunction(enable_clang_tidy)


add_subdirectory(fs)
add_subdirectory(test)
