add_executable(naxer main.cpp)
target_link_libraries(naxer pthread nbre_common nbre_core nbre_jit
  nbre_fs boost_system boost_date_time rocksdb bz2 z zstd snappy lz4
  ${protobuf_lib} ${PROJECT_SOURCE_DIR}/lib/lib/libc++.so.1 gflags)
