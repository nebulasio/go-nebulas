add_executable(test_ipc main.cpp
  ipc_instance.cpp
  itest_example.cpp
  )

target_link_libraries(test_ipc ${common_libs} ${fs_libs})

add_test(NAME test_ipc COMMAND test_ipc)
