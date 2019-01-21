project(FunctionFlow)
cmake_minimum_required(VERSION 2.6)
message(STATUS "This is BINARY dir" ${PROJECT_BINARY_DIR})
message(STATUS "This is SOURCE dir" ${PROJECT_SOURCE_DIR})
add_definitions(-std=c++11)
add_definitions(-stdlib=libc++)

if(Release)
	message(STATUS "This is a Release version!")
	add_definitions(-O2)
	add_definitions(-DNDEBUG)
else(Release)
	message(STATUS "This is a Debug version!")
	add_definitions(-g)
        add_definitions(-O0)
        add_definitions(-DFUNCTION_FLOW_DEBUG)
	include_directories(${PROJECT_SOURCE_DIR}/libs)
	if(NOT EXISTS ${PROJECT_SOURCE_DIR}/libs/ff)
		message(STATUS "require ff, start downloading...")
		file(DOWNLOAD https://github.com/AthrunArthur/ff/archive/master.zip ${PROJECT_SOURCE_DIR}/libs/ff.zip)
		exec_program(unzip ${PROJECT_SOURCE_DIR}/libs ARGS ff.zip */ff/* -d ffcommon)
		exec_program(cp ${PROJECT_SOURCE_DIR}/libs ARGS -r ffcommon/ff-master/ff/ ff/)
		exec_program(rm ${PROJECT_SOURCE_DIR}/libs ARGS -rf ffcommon/)
	endif(NOT EXISTS ${PROJECT_SOURCE_DIR}/libs/ff)
endif(Release)

#add_definitions(-DUSING_SPIN_STEAL_QUEUE)
#add_definitions(-DUSING_MUTEX_STEAL_QUEUE)
add_definitions(-DUSING_WORK_STEALING_QUEUE)
set(LIBRARY_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/bin)


include_directories(${PROJECT_SOURCE_DIR}/ff/include)
#link_directories(${PROJECT_SOURCE_DIR}/ff/bin)

set(ff_src ff/src/common.cpp
  ff/src/rtcmn.cpp
  ff/src/runtime.cpp
  ff/src/wait.cpp)

add_library(ff SHARED ${ff_src})
add_library(ff_s STATIC ${ff_src})
target_link_libraries(ff pthread)
target_link_libraries(ff_s pthread)
set_target_properties(ff_s PROPERTIES OUTPUT_NAME "ff")

install(TARGETS ff ff_s
RUNTIME DESTINATION bin
LIBRARY DESTINATION lib
ARCHIVE DESTINATION lib)

install(DIRECTORY ${PROJECT_SOURCE_DIR}/ff/include/ff DESTINATION include)
