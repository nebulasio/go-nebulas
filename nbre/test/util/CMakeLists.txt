add_executable(test_util main.cpp
  gtest_currency.cpp)

target_link_libraries(test_util nbre_rt ${gtest_lib})
