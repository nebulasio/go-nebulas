include(GoogleTest)

set(common_libs ${common_libs} gtest)
#add_subdirectory(link_example)
add_subdirectory(common)
add_subdirectory(fs)
add_subdirectory(core)
