add_library(nbre_fs rocksdb_storage.cpp)

target_link_libraries(nbre_fs rocksdb)

enable_clang_tidy(nbre_fs)
