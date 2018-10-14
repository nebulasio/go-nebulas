add_executable(test_ipc main.cpp
  ipc_instance.cpp
  itest_example.cpp
  itest_bookkeeper.cpp
  itest_session.cpp
  itest_shm_service.cpp
  )

target_link_libraries(test_ipc nbre_rt)

add_test(NAME test_ipc COMMAND test_ipc)
