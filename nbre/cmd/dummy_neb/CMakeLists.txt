include_directories(${PROJECT_SOURCE_DIR}/cmd/dummy_neb/)
file(GLOB_RECURSE dummy_src *.cpp)
add_executable(dummy_neb ${dummy_src})
target_link_libraries(dummy_neb nbre_rt)
