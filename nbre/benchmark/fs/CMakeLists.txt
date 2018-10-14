
add_executable(benchmark_fs main.cpp rocksdb_storage.cpp blockchain.cpp nbre_storage.cpp)
target_link_libraries(benchmark_fs nbre_rt nbre_benchmark_instances)
