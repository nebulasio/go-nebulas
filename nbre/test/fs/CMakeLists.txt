add_executable(test_fs main.cpp
  gtest_util.cpp
  gtest_rocksdb_storage.cpp
  gtest_blockchain.cpp
  ir_manager/gtest_ir_manager.cpp
  blockchain/gtest_trie.cpp
  #blockchain/gtest_transaction_db.cpp
  #blockchain/gtest_account_db.cpp
  )

target_link_libraries(test_fs nbre_rt ${gtest_lib})

#gtest_discover_tests(test_fs)
add_test(NAME test_fs
  COMMAND test_fs)

add_executable(rocksdb_read rocksdb_read.cpp)
add_executable(rocksdb_write rocksdb_write.cpp)

target_link_libraries(rocksdb_read nbre_rt )
target_link_libraries(rocksdb_write nbre_rt )

