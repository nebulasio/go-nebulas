add_executable(test_core main.cpp gtest_ir_warden.cpp)

target_link_libraries(test_core nbre_core ${common_libs} ${fs_libs})
add_test(NAME test_core
  COMMAND test_core)

add_executable(ir_warden test_ir_warden.cpp)
target_link_libraries(ir_warden nbre_core ${common_libs} ${fs_libs})

add_test(NAME test_ir_warden
  COMMAND ir_warden)

