include(GoogleTest)

#set(common_libs ${common_libs} gtest)
#add_subdirectory(link_example)
set(gtest_lib ${PROJECT_SOURCE_DIR}/lib/lib/libgtest.a)
add_subdirectory(common)
add_subdirectory(fs)
add_subdirectory(core)
add_subdirectory(jit)
