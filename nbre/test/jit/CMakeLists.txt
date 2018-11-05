add_executable(test_jit main.cpp gtest_jit_driver.cpp)

target_link_libraries(test_jit nbre_rt gtest)
gtest_discover_tests(test_jit)
add_test(NAME test_jit
  COMMAND test_jit)

add_test(NAME test_jit_driver
  COMMAND jit_driver)

