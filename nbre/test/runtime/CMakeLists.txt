add_executable(test_runtime main.cpp
  nr/gtest_graph.cpp
  nr/gtest_algo.cpp
  nr/gtest_nebulas_rank.cpp
  )

target_link_libraries(test_runtime nbre_rt ${gtest_lib})
