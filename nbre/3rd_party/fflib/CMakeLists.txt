cmake_minimum_required(VERSION 2.8)
project(fflib)

add_definitions(-std=c++14)
add_definitions(-O2)
add_definitions(-DNDEBUG)
#add_definitions(-g)
#add_definitions(-DPROTO_BUF_SUPPORT)


include_directories(/usr/local/include)
include_directories(${PROJECT_SOURCE_DIR}/include)
find_package(Boost REQUIRED)
include_directories(${Boost_INCLUDE_DIRS})
link_directories(${Boost_LIBRARY_DIRS})
link_directories(/usr/local/lib)

set(LIBRARY_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/lib)


file(GLOB_RECURSE functionflow_src src/functionflow/*.cpp)
file(GLOB_RECURSE net_src src/net/*.cpp)
file(GLOB_RECURSE public_header_src include/ff/*.h)

set(lib_list glog gflags)

find_package(Protobuf QUIET)
if(PROTOBUF_FOUND)
  #message(STATUS "found protobuf, and include protobuf support")
  #add_definitions(-DPROTO_BUF_SUPPORT)
  #include_directories(${PROTOBUF_INCLUDE_DIRS})
  #set(src_list ${src_list} extension/protobuf/proto_buf_wrapper_pkg.cpp
  #  extension/protobuf/protobuf_pkg_hub.cpp)
  #set(lib_list ${lib_list} ${PROTOBUF_LIBRARIES})
else()
  message(STATUS "not found protobuf, gonna ignore protobuf support")
endif()

add_library(functionflow_dynamic SHARED ${functionflow_src})
add_library(functionflow_static STATIC ${functionflow_src})
set_target_properties(functionflow_dynamic PROPERTIES OUTPUT_NAME "ff_functionflow")
set_target_properties(functionflow_static  PROPERTIES OUTPUT_NAME "ff_functionflow")

add_library(net_dynamic SHARED ${net_src})
add_library(net_static STATIC ${net_src})
target_link_libraries(net_dynamic boost_program_options boost_system)
target_link_libraries(net_static boost_program_options boost_system)
set_target_properties(net_dynamic PROPERTIES OUTPUT_NAME "ff_net")
set_target_properties(net_static  PROPERTIES OUTPUT_NAME "ff_net")

set_target_properties(net_static PROPERTIES PUBLIC_HEADER "${public_header_src}")
install(TARGETS functionflow_static functionflow_dynamic net_static net_dynamic
  LIBRARY DESTINATION "lib"
  ARCHIVE DESTINATION "lib")
install(
  DIRECTORY include/ff DESTINATION "include")



if(Release)
else()
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${PROJECT_SOURCE_DIR}/bin/)
link_directories(${PROJECT_SOURCE_DIR}/lib)
add_subdirectory(example)
endif()
