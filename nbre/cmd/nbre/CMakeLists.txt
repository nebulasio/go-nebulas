add_executable(nbre main.cpp)
target_link_libraries(nbre nbre_rt ${PROJECT_SOURCE_DIR}/lib/lib/libgflags.a)
set_target_properties(nbre PROPERTIES LINK_FLAGS "-rdynamic")
