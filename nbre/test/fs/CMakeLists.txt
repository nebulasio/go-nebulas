add_executable(test_fs main.cpp
  gtest_util.cpp
  gtest_rocksdb_storage.cpp
  gtest_blockchain.cpp
  gtest_nbre_storage.cpp
  )

target_link_libraries(test_fs ${common_libs} ${fs_libs})

gtest_discover_tests(test_fs)
add_test(NAME test_fs
  COMMAND test_fs)

add_test(NAME test_blockchain COMMAND blockchain)
add_test(NAME test_nbre_storage COMMAND nbre_storage)
add_test(NAME test_rocksdb_storage COMMAND rocksdb_storage)

