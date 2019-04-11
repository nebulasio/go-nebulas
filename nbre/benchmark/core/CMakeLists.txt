
add_executable(benchmark_core main.cpp ir_warden.cpp)
target_link_libraries(benchmark_core nbre_rt nbre_benchmark_instances)
