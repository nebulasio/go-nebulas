add_executable(test_fs main.cpp
  gtest_util.cpp
  gtest_storage.cpp)

target_link_libraries(test_fs gtest pthread nbre_common nbre_fs ${Boost_LIBRARIES}
  rocksdb bz2 z zstd snappy lz4 ${protobuf_lib} ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so.1)

add_test(NAME test_fs
  COMMAND test_fs)

add_executable(blockchain test_blockchain.cpp)
target_link_libraries(blockchain pthread nbre_common nbre_fs ${Boost_LIBRARIES}
  rocksdb bz2 z zstd snappy lz4 ${protobuf_lib} ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so.1 gflags)

add_executable(nbre_storage test_nbre_storage.cpp)
target_link_libraries(nbre_storage pthread nbre_common nbre_fs ${Boost_LIBRARIES}
  rocksdb bz2 z zstd snappy lz4 ${protobuf_lib} ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so.1 gflags)


