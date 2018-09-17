add_executable(test_fs main.cpp
  gtest_util.cpp
  gtest_storage.cpp)

target_link_libraries(test_fs gtest pthread nbre_common nbre_fs ${Boost_LIBRARIES}
  rocksdb bz2 z zstd snappy lz4 ${protobuf_lib})

add_test(NAME test_fs
  COMMAND test_fs)


