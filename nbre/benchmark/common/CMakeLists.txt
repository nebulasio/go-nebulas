
#add_executable(benchmark_example main.cpp example.cpp)
add_executable(benchmark_math main.cpp math.cpp )
#target_link_libraries(benchmark_example nbre_rt nbre_benchmark_instances)
target_link_libraries(benchmark_math nbre_rt nbre_benchmark_instances)

add_executable(benchmark_seralize main.cpp serialize.cpp)
target_link_libraries(benchmark_seralize nbre_rt nbre_benchmark_instances)
