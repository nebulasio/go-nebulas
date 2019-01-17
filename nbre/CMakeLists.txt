cmake_minimum_required(VERSION 3.12)
project(nebulas-blockchain-runtime-environment)
add_definitions(-std=c++14)
add_definitions(-stdlib=libc++)

if(Release)
  message(STATUS "This is a Release version")
  add_definitions(-O2)
  add_definitions(-DNBRE_PATH=${PROJECT_SOURCE_DIR})
else(Release)
  add_definitions(-DNBRE_PATH=${PROJECT_SOURCE_DIR})
endif(Release)

find_program(
  CLANG_TIDY_EXE
  NAMES "clang-tidy"
  DOC "Path to clang-tidy executable"
)

if(NOT CLANG_TIDY_EXE)
  message(STATUS "clang-tidy not found.")
else()
  if(Release)
  else()
    message(STATUS "clang-tidy found: ${CLANG_TIDY_EXE}")
    set(DO_CLANG_TIDY "${CLANG_TIDY_EXE}" "-checks=*,-clang-analyzer-alpha.*")
  endif()
endif()

find_package(LLVM REQUIRED CONFIG)
include_directories(${LLVM_INCLUDE_DIRS})
add_definitions(${LLVM_DEFINITIONS})

include_directories(${PROJECT_SOURCE_DIR})
include_directories(${PROJECT_SOURCE_DIR}/lib/include)
link_directories(${PROJECT_SOURCE_DIR}/lib/lib)
message(STATUS 'lib dir ' ${PROJECT_SOURCE_DIR}/lib/lib)

set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${PROJECT_SOURCE_DIR}/bin/)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${PROJECT_SOURCE_DIR}/lib/lib/)
set(LIBRARY_OUTPUT_DIRECTORY ${PROJECT_SOURCE_DIR}/lib/lib/)

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


llvm_map_components_to_libnames(llvm_libs
  core
  executionengine
  interpreter
  mc mcjit
  support
  nativecodegen
  irreader
  orcjit
  )

set(jit_libs nbre_jit ${llvm_libs})

set(fs_libs nbre_fs nbre_fs_proto rocksdb bz2 z zstd snappy lz4)

if(UNIX AND NOT APPLE)
  set(protobuf_lib ${PROJECT_SOURCE_DIR}/lib/lib/libprotobuf.so)
  set(cpp_lib ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so)
  set(Boost_LIBRARIES
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_filesystem.so
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_system.so
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_date_time.so
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_graph.so
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_program_options.so)
  set(extra_libs
    ${PROJECT_SOURCE_DIR}/lib/lib/libglog.a
    ${PROJECT_SOURCE_DIR}/lib/lib/librocksdb.so.5
    rt
    bz2 z zstd snappy lz4
    )
elseif(UNIX)
  set(common_libs ${common_libs} ${PROJECT_SOURCE_DIR}/lib/lib/libgflags.a)
  set(protobuf_lib ${PROJECT_SOURCE_DIR}/lib/lib/libprotobuf.dylib)
  set(cpp_lib ${PROJECT_SOURCE_DIR}/lib/lib/libc++.dylib)
  set(Boost_LIBRARIES
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_filesystem.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_system.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_date_time.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_graph.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_program_options.dylib)
  set(extra_libs
    ${PROJECT_SOURCE_DIR}/lib/lib/libgflags.a
    ${PROJECT_SOURCE_DIR}/lib/lib/libglog.a
    ${PROJECT_SOURCE_DIR}/lib/lib/libbz2.1.dylib
    rocksdb z zstd snappy lz4) #TODO add rocksdb here
endif()
set(common_libs nbre_common pthread ${common_libs} ${protobuf_lib} ${cpp_lib} ${Boost_LIBRARIES})


message(STATUS "nbre src " ${nbre_src})
file(GLOB nbre_jit_src jit/*.cpp)
file(GLOB_RECURSE nbre_rt_src runtime/*.cpp)
file(GLOB_RECURSE nbre_common_src common/*.cpp)
file(GLOB_RECURSE nbre_fs_src fs/*.cpp)
file(GLOB_RECURSE nbre_fs_proto_src fs/proto/*.pb.cc)
file(GLOB_RECURSE nbre_core_src core/*.cpp)

add_library(nbre_rt SHARED ${nbre_jit_src} ${nbre_rt_src} ${nbre_common_src} ${nbre_fs_src} ${nbre_fs_proto_src} ${nbre_core_src})
target_link_libraries(nbre_rt ${llvm_libs} ${protobuf_lib} ${cpp_lib} ${Boost_LIBRARIES} ${extra_libs} ff softfloat c++)

if(Release)
else()
add_subdirectory(test)
endif()
#add_subdirectory(benchmark)

if(EXISTS "${PROJECT_SOURCE_DIR}/cmd/CMakeLists.txt")
  add_subdirectory(cmd)
endif()
