add_executable(naxer main.cpp)
target_link_libraries(naxer nbre_core ${jit_libs} ${common_libs} ${fs_libs} gflags)
