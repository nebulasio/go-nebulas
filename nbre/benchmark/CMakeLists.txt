
add_library(nbre_benchmark_instances benchmark_instances.cpp)
set(common_libs ${common_libs}
  boost_date_time
  boost_program_options
  nbre_benchmark_instances
)

add_subdirectory(common)
#add_subdirectory(fs)
add_subdirectory(core)
