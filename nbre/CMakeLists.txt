cmake_minimum_required(VERSION 3.12)
project(nebulas-blockchain-runtime-environment)

add_definitions(-std=c++14)
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

find_package(Boost 1.67.0 REQUIRED COMPONENTS filesystem system)

include_directories(${PROJECT_SOURCE_DIR})
include_directories(${PROJECT_SOURCE_DIR}/lib/include)
link_directories(${PROJECT_SOURCE_DIR}/lib/lib)
message(STATUS 'lib dir ' ${PROJECT_SOURCE_DIR}/lib/lib)

set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${PROJECT_SOURCE_DIR}/bin/)

function(enable_clang_tidy target_name)
  set_target_properties(
    ${target_name} PROPERTIES
    CXX_STANDARD 14
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

enable_testing()

set(protobuf_lib ${PROJECT_SOURCE_DIR}/lib/lib/libprotobuf.so)

add_subdirectory(common)
add_subdirectory(fs)
add_subdirectory(test)

if(EXISTS "${PROJECT_SOURCE_DIR}/cmd/CMakeLists.txt")
  add_subdirectory(cmd)
endif()
