add_executable(gtest_core main.cpp gtest_ir_warden.cpp)

target_link_libraries(gtest_core nbre_rt ${gtest_lib})
#gtest_discover_tests(test_core)
add_test(NAME gtest_core
  COMMAND gtest_core)

#add_subdirectory(neb_ipc)

