cmake_minimum_required(VERSION 3.12)
project(nebulas-blockchain-runtime-environment)
add_definitions(-std=c++14)

if(Release)
  message(STATUS "This is a Release version")
  add_definitions(-O2)
  add_definitions(-DNBRE_PATH=${PROJECT_SOURCE_DIR})
  add_definitions(-DNDEBUG)
else(Release)
  add_definitions(-O2)
  add_definitions(-g)
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

set(CMAKE_SKIP_RPATH true)
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

if(CMAKE_BUILD_TYPE STREQUAL "coverage" OR CODE_COVERAGE)
  message("Building with llvm Code Coverage Tools")
  # set Flags
  set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fprofile-instr-generate -fcoverage-mapping")
  set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fprofile-instr-generate -fcoverage-mapping")
endif()

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

set(fs_libs nbre_fs nbre_fs_proto bz2 z zstd snappy lz4)

if(UNIX AND NOT APPLE)
  set(common_libs ${common_libs}
    ${PROJECT_SOURCE_DIR}/lib/lib/libgflags.so
    ${PROJECT_SOURCE_DIR}/lib/lib/libcryptopp.so)
  set(protobuf_lib ${PROJECT_SOURCE_DIR}/lib/lib/libprotobuf.so)
  set(cpp_lib ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so)
  set(Boost_LIBRARIES
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_filesystem.so
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_system.so
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_date_time.so
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_graph.so
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_thread.so
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_program_options.so)
  set(extra_libs
    ${PROJECT_SOURCE_DIR}/lib/lib/libglog.so
    ${PROJECT_SOURCE_DIR}/lib/lib/librocksdb.so.5
    rt
    bz2 z zstd snappy lz4
    )
elseif(UNIX)
  set(common_libs ${common_libs}
    ${PROJECT_SOURCE_DIR}/lib/lib/libgflags.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libcryptopp.dylib)
  set(protobuf_lib ${PROJECT_SOURCE_DIR}/lib/lib/libprotobuf.dylib)
  set(cpp_lib ${PROJECT_SOURCE_DIR}/lib/lib/libc++.dylib)
  set(Boost_LIBRARIES
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_filesystem.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_system.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_date_time.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_graph.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_thread.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libboost_program_options.dylib)
  set(extra_libs
    ${PROJECT_SOURCE_DIR}/lib/lib/libgflags.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libglog.dylib
    ${PROJECT_SOURCE_DIR}/lib/lib/libbz2.1.dylib
    rocksdb z zstd snappy lz4) #TODO add rocksdb here
endif()
set(common_libs nbre_common pthread
  ${common_libs}
  ${protobuf_lib} ${cpp_lib} ${Boost_LIBRARIES}
  )


message(STATUS "nbre src " ${nbre_src})
file(GLOB nbre_jit_src jit/*.cpp)
file(GLOB_RECURSE nbre_rt_src runtime/*.cpp)
file(GLOB_RECURSE nbre_common_src common/*.cpp)
file(GLOB_RECURSE nbre_fs_src fs/*.cpp)
file(GLOB_RECURSE nbre_fs_proto_src fs/proto/*.pb.cc)
file(GLOB_RECURSE nbre_core_src core/*.cpp)
file(GLOB_RECURSE nbre_util_src util/*.cpp)
file(GLOB_RECURSE nbre_crypto_src crypto/*.cpp)

list(FILTER nbre_core_src EXCLUDE REGEX "core/driver.cpp")
list(FILTER nbre_core_src EXCLUDE REGEX "core/neb_ipc*")
list(FILTER nbre_common_src EXCLUDE REGEX "common/ipc*")

add_library(nbre_rt SHARED
  ${nbre_jit_src}
  ${nbre_rt_src}
  ${nbre_common_src}
  ${nbre_fs_src}
  ${nbre_fs_proto_src}
  ${nbre_core_src}
  ${nbre_util_src}
  ${nbre_crypto_src}
  )
if(UNIX AND NOT APPLE)
  target_link_libraries(nbre_rt ${PROJECT_SOURCE_DIR}/lib/lib/libgflags.so ${llvm_libs} ${protobuf_lib} ${cpp_lib} ${Boost_LIBRARIES} ${extra_libs} ff_functionflow ff_net softfloat c++ cryptopp tcmalloc)
elseif(UNIX)
  target_link_libraries(nbre_rt ${PROJECT_SOURCE_DIR}/lib/lib/libgflags.dylib ${llvm_libs} ${protobuf_lib} ${cpp_lib} ${Boost_LIBRARIES} ${extra_libs} ff_functionflow ff_net softfloat c++ cryptopp tcmalloc)
endif()

if(Release)
else()
  add_subdirectory(test)
endif()
if(UNIX AND NOT APPLE)
  add_subdirectory(benchmark)
endif()


file(GLOB_RECURSE DEPLOY_FILES_AND_DIRS_common common/*.h common/*.hpp)
file(GLOB_RECURSE DEPLOY_FILES_AND_DIRS_core core/*.h core/*.hpp)
file(GLOB_RECURSE DEPLOY_FILES_AND_DIRS_rt rt/*.h rt/*.hpp)
file(GLOB_RECURSE DEPLOY_FILES_AND_DIRS_fs fs/*.h fs/*.hpp)
file(GLOB_RECURSE DEPLOY_FILES_AND_DIRS_util util/*.h util/*.hpp)
file(GLOB_RECURSE DEPLOY_FILES_AND_DIRS_crypto crypto/*.h crypto/*.hpp)
file(GLOB_RECURSE DEPLOY_FILES_AND_DIRS_runtime runtime/*.h runtime/*.hpp)

set(DEPLOY_FILES_AND_DIRS ${DEPLOY_FILES_AND_DIRS_common}
  ${DEPLOY_FILES_AND_DIRS_core}
  ${DEPLOY_FILES_AND_DIRS_rt}
  ${DEPLOY_FILES_AND_DIRS_fs}
  ${DEPLOY_FILES_AND_DIRS_util}
  ${DEPLOY_FILES_AND_DIRS_crypto}
  ${DEPLOY_FILES_AND_DIRS_runtime})

MACRO(INSTALL_HEADERS_WITH_DIRECTORY HEADER_LIST)

  FOREACH(HEADER ${${HEADER_LIST}})
    #STRING(REGEX MATCH "(.\\\*)\\\[/\\\]" DIR ${HEADER})
    STRING(REPLACE "${PROJECT_SOURCE_DIR}/" "" RPATH ${HEADER})
    STRING(REGEX MATCH "(.*)/"  DIR ${RPATH})
    INSTALL(FILES ${HEADER} DESTINATION ${PROJECT_SOURCE_DIR}/lib/include/${DIR})
  ENDFOREACH(HEADER)

ENDMACRO(INSTALL_HEADERS_WITH_DIRECTORY)

INSTALL_HEADERS_WITH_DIRECTORY(DEPLOY_FILES_AND_DIRS)

if(EXISTS "${PROJECT_SOURCE_DIR}/cmd/CMakeLists.txt")
  add_subdirectory(cmd)
endif()
