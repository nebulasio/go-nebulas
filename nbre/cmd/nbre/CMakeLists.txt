add_executable(nbre main.cpp)
target_link_libraries(nbre nbre_core ${jit_libs} ${common_libs} ${fs_libs})
