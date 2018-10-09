add_executable(test_fs main.cpp
  gtest_util.cpp
  gtest_rocksdb_storage.cpp
  gtest_blockchain.cpp
  gtest_nbre_storage.cpp
  )

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

add_executable(rocksdb_storage test_rocksdb_storage.cpp)
target_link_libraries(rocksdb_storage pthread nbre_common nbre_fs ${Boost_LIBRARIES}
  rocksdb bz2 z zstd snappy lz4 ${protobuf_lib} ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so.1 gflags)

add_test(NAME test_blockchain COMMAND blockchain)
add_test(NAME test_nbre_storage COMMAND nbre_storage)
add_test(NAME test_rocksdb_storage COMMAND rocksdb_storage)

