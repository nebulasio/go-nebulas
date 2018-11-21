add_library(nbre_fs
  rocksdb_storage.cpp
  util.cpp
  blockchain.cpp
  manager/nbre_storage.cpp
  manager/flag_storage.cpp
  api/nbre_api.cpp
  trie/trie.cpp
)

add_library(nbre_fs_proto
  proto/block.pb.cc
  proto/dag.pb.cc
  proto/dynasty.pb.cc
  proto/genesis.pb.cc
  proto/state.pb.cc
  proto/ir.pb.cc)

target_link_libraries(nbre_fs_proto ${protobuf_lib})
target_link_libraries(nbre_fs rocksdb nbre_common nbre_fs_proto ${protobuf_lib})

enable_clang_tidy(nbre_fs)
