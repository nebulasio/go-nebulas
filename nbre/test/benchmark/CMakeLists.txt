
add_executable(test_benchmark test_benchmark.cpp account_db_v2.cpp nebulas_rank_v2.cpp)
target_link_libraries(test_benchmark nbre_rt)
