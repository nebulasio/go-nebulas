add_executable(gtest_util main.cpp
  gtest_currency.cpp)

target_link_libraries(gtest_util nbre_rt ${gtest_lib})
