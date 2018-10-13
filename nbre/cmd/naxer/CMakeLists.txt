add_executable(naxer main.cpp)
target_link_libraries(naxer nbre_rt)
set_target_properties(naxer PROPERTIES LINK_FLAGS "-rdynamic")
