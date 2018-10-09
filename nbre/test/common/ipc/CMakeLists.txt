add_executable(test_ipc main.cpp
  ipc_instance.cpp
  itest_example.cpp
  )

target_link_libraries(test_ipc pthread nbre_common rt glog boost_system boost_filesystem nbre_fs rocksdb bz2 z zstd snappy lz4 ${protobuf_lib} ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so.1)

add_test(NAME test_ipc COMMAND test_ipc)
