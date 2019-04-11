
add_executable(trace_rocksdb trace_rocksdb.cpp)
target_link_libraries(trace_rocksdb nbre_rt)

add_executable(blockchain blockchain.cpp)
target_link_libraries(blockchain nbre_rt)

add_executable(auth_table auth_table.cpp)
target_link_libraries(auth_table nbre_rt)

add_executable(rocksdb_storage rocksdb_storage.cpp)
target_link_libraries(rocksdb_storage nbre_rt)

add_executable(rocksdb_rm rocksdb_rm.cpp)
target_link_libraries(rocksdb_rm nbre_rt)

