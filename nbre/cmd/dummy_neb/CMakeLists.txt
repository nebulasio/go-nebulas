include_directories(${PROJECT_SOURCE_DIR}/cmd/dummy_neb/)
file(GLOB_RECURSE dummy_src *.cpp)
list(FILTER dummy_src EXCLUDE REGEX "cli/main.cpp")

add_executable(dummy_neb ${dummy_src})
target_link_libraries(dummy_neb nbre_rt)

add_executable(dummy cli/main.cpp)
target_link_libraries(dummy nbre_rt)
