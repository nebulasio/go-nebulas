include_directories(${PROJECT_SOURCE_DIR}/cmd/dummy_neb/)
add_executable(dummy_neb main.cpp)
target_link_libraries(dummy_neb nbre_rt)
