add_library(nbre_fs rocksdb_storage.cpp util.cpp
  proto/block.pb.cc
  proto/dag.pb.cc
  proto/dynasty.pb.cc
  proto/genesis.pb.cc
  proto/state.pb.cc)

target_link_libraries(nbre_fs rocksdb)

enable_clang_tidy(nbre_fs)
