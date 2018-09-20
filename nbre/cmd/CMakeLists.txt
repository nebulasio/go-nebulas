
add_executable(blockchain blockchain.cpp)
target_link_libraries(blockchain pthread nbre_common nbre_fs ${Boost_LIBRARIES}
  rocksdb bz2 z zstd snappy lz4 ${protobuf_lib} ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so.1 gflags)

