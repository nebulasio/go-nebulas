
add_executable(test_crypto main.cpp
  gtest_hash.cpp)

target_link_libraries(test_crypto nbre_rt ${gtest_lib})

#gtest_discover_tests(test_crypto)

add_test(NAME test_crypto
  COMMAND test_crypto)

set_tests_properties(
  test_crypto PROPERTIES
  ENVIRONMENT "LD_LIBRARY_PATH=${PROJECT_SOURCE_DIR}/lib/lib"
)
