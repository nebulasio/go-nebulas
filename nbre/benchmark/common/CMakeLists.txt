add_library(nbre_benchmark_common benchmark_instances.cpp)

add_executable(benchmark_example main.cpp example.cpp)
target_link_libraries(benchmark_example ${common_libs} nbre_benchmark_common)
