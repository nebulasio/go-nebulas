add_executable(jit main.cpp)
target_link_libraries(jit nbre_rt)
set_target_properties(naxer PROPERTIES LINK_FLAGS "-rdynamic")
