add_executable(dummy_neb main.cpp)
target_link_libraries(dummy_neb nbre_core ${jit_libs} ${common_libs} ${fs_libs})
