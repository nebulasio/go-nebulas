add_executable(test_neb_ipc_server server.cpp)

target_link_libraries(test_neb_ipc_server nbre_rt)

add_executable(test_neb_ipc_client client.cpp)

target_link_libraries(test_neb_ipc_client nbre_rt)
