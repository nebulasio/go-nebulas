add_executable(test_core main.cpp gtest_ir_warden.cpp)

target_link_libraries(test_core gtest pthread nbre_core nbre_fs nbre_common ${Boost_LIBRARIES}
  rocksdb bz2 z zstd snappy lz4 ${protobuf_lib} ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so.1)
add_test(NAME test_core
  COMMAND test_core)

add_executable(ir_warden test_ir_warden.cpp)
target_link_libraries(ir_warden pthread nbre_core nbre_fs nbre_common ${Boost_LIBRARIES}
  rocksdb bz2 z zstd snappy lz4 ${protobuf_lib} ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so.1 gflags)

add_test(NAME test_ir_warden
  COMMAND ir_warden)

