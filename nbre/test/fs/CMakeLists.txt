add_executable(test_fs main.cpp
  gtest_util.cpp
  gtest_rocksdb_storage.cpp
  gtest_blockchain.cpp
  gtest_nbre_storage.cpp
  gtest_trie.cpp
  gtest_transaction_db.cpp
  gtest_account_db.cpp
  )

target_link_libraries(test_fs  nbre_rt gtest)

gtest_discover_tests(test_fs)
add_test(NAME test_fs
  COMMAND test_fs)

