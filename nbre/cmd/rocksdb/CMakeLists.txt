
add_executable(trace_rocksdb main.cpp)
target_link_libraries(trace_rocksdb ${common_libs} ${fs_libs} gflags)

