
add_executable(test_storage_holder main.cpp
  storage_holder.cpp
  )

target_link_libraries(test_storage_holder nbre_rt ${gtest_lib})

add_executable(toy toy.cpp)
target_link_libraries(toy nbre_rt)

add_executable(hash hash.cpp)
target_link_libraries(hash nbre_rt)

