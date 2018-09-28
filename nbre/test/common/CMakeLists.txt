add_executable(test_common main.cpp gtest_util.cpp gtest_util_json.cpp gtest_util_conf.cpp)
target_link_libraries(test_common gtest pthread nbre_common glog)

add_test(NAME test_common
  COMMAND test_common)


