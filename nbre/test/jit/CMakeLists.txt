#add_executable(test_jit main.cpp gtest_jit_driver.cpp gtest_orclazy_jit.cpp)
add_executable(gtest_jit main.cpp
  ir/gen_ir.cpp
  gtest_jit_driver.cpp
  gtest_cpp_ir.cpp
  gtest_jit_mangled_entry_point.cpp
  )
target_link_libraries(gtest_jit nbre_rt ${gtest_lib})
#gtest_discover_tests(test_jit)
add_test(NAME gtest_jit
  COMMAND gtest_jit)

add_executable(test_jit_driver test_jit_driver.cpp)
target_link_libraries(test_jit_driver nbre_rt)

add_executable(test_jit_mangled_entry_point test_jit_mangled_entry_point.cpp)
target_link_libraries(test_jit_mangled_entry_point nbre_rt)
set_target_properties(test_jit_mangled_entry_point PROPERTIES LINK_FLAGS "-rdynamic")
