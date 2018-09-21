add_executable(test_core main.cpp gtest_ir_warden.cpp)

target_link_libraries(test_core gtest pthread nbre_core nbre_common glog ${Boost_LIBRARIES})
add_test(NAME test_core
  COMMAND test_core)
