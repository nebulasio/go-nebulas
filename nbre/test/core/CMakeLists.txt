add_executable(test_core main.cpp gtest_ir_warden.cpp)

target_link_libraries(test_core nbre_rt ${gtest_lib})
#gtest_discover_tests(test_core)
add_test(NAME test_core
  COMMAND test_core)

add_test(NAME test_ir_warden
  COMMAND ir_warden)

#add_subdirectory(neb_ipc)

