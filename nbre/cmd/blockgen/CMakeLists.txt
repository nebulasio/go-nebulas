add_executable(blockgen main.cpp)
target_link_libraries(blockgen ${common_libs} ${fs_libs} boost_program_options)

