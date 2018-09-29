add_executable(blockgen main.cpp)
target_link_libraries(blockgen pthread nbre_common nbre_fs ${Boost_LIBRARIES}
  rocksdb bz2 z zstd snappy lz4 ${protobuf_lib}
  ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so.1 gflags
  nbre_fs_proto boost_program_options)

