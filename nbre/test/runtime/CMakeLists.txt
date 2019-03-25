add_executable(test_runtime main.cpp
  nr/gtest_graph.cpp
  nr/gtest_algo.cpp
  nr/gtest_nebulas_rank.cpp
  dip/gtest_dip_reward.cpp
  nr/gtest_decycle.cpp
  )

target_link_libraries(test_runtime nbre_rt ${gtest_lib})

add_executable(test_algo nr/test_algo.cpp)
target_link_libraries(test_algo nbre_rt)

