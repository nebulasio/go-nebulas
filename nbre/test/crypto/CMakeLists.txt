
add_executable(gtest_crypto main.cpp
  gtest_hash.cpp)

target_link_libraries(gtest_crypto nbre_rt ${gtest_lib})

#gtest_discover_tests(test_crypto)

add_test(NAME gtest_crypto
  COMMAND gtest_crypto)

set_tests_properties(
  gtest_crypto PROPERTIES
  ENVIRONMENT "LD_LIBRARY_PATH=${PROJECT_SOURCE_DIR}/lib/lib"
)
