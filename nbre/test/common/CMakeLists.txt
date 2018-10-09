add_executable(test_common main.cpp
  gtest_util.cpp
  gtest_util_json.cpp
  gtest_util_conf.cpp
  )

target_link_libraries(test_common ${common_libs})

add_test(NAME test_common
  COMMAND test_common)

add_subdirectory(ipc)


