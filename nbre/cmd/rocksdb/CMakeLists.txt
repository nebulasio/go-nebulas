
add_executable(trace_rocksdb trace_rocksdb.cpp)
target_link_libraries(trace_rocksdb nbre_rt)

add_executable(blockchain blockchain.cpp)
target_link_libraries(blockchain nbre_rt)

