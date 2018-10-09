cmake_minimum_required(VERSION 3.12)
project(nebulas-blockchain-runtime-environment)

#add_definitions(-DGOOGLE_PROTOBUF_NO_RTTI)
add_definitions(-std=c++14)
#add_definitions(-fno-rtti)
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

set(Boost_LIBRARIES boost_filesystem boost_system boost_date_time)

include_directories(${PROJECT_SOURCE_DIR})
include_directories(${PROJECT_SOURCE_DIR}/lib/include)
link_directories(${PROJECT_SOURCE_DIR}/lib/lib)
message(STATUS 'lib dir ' ${PROJECT_SOURCE_DIR}/lib/lib)

set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${PROJECT_SOURCE_DIR}/bin/)
set(LIBRARY_OUTPUT_DIRECTORY ${PROJECT_SOURCE_DIR}/bin/)

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

set(common_libs nbre_common
  ${protobuf_lib}
  ${Boost_LIBRARIES}
  pthread
  glog
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

set(jit_libs nbre_jit ${llvm_libs})

set(fs_libs nbre_fs nbre_fs_proto rocksdb bz2 z zstd snappy lz4)

if(UNIX AND NOT APPLE)
  set(common_libs ${common_libs} rt)
  set(protobuf_lib ${PROJECT_SOURCE_DIR}/lib/lib/libprotobuf.so)
  set(cpp_lib ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so)
elseif(UNIX)
  set(protobuf_lib ${PROJECT_SOURCE_DIR}/lib/lib/libprotobuf.dylib)
  set(cpp_lib ${PROJECT_SOURCE_DIR}/lib/lib/libc++.dylib)
endif()
set(common_libs ${common_libs} ${protobuf_lib} ${cpp_lib})

add_subdirectory(common)
add_subdirectory(fs)
add_subdirectory(test)
add_subdirectory(core)
add_subdirectory(jit)
add_subdirectory(benchmark)

if(EXISTS "${PROJECT_SOURCE_DIR}/cmd/CMakeLists.txt")
  add_subdirectory(cmd)
endif()
