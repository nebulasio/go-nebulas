add_executable(nbre main.cpp)
target_link_libraries(nbre nbre_rt)
set_target_properties(nbre PROPERTIES LINK_FLAGS "-rdynamic")
