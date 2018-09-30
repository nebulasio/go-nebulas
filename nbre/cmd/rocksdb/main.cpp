
#include "common/util/version.h"
#include "fs/nbre_storage.h"
#include "fs/proto/block.pb.h"
#include "fs/proto/ir.pb.h"
#include "fs/rocksdb_storage.h"
#include "fs/util.h"
#include <gflags/gflags.h>

DEFINE_string(db_path, "", "rocksdb file path");
DEFINE_string(func_type, "keys", "call back function type");
DEFINE_string(key, "nr", "rocksdb key");
DEFINE_int64(max_height, 0, "nbre max height set");
DEFINE_int64(height, 0, "nbre height");

int main(int argc, char *argv[]) {
  google::ParseCommandLineFlags(&argc, &argv, true);
  std::string db_path = FLAGS_db_path;
  std::string func_type = FLAGS_func_type;
  neb::block_height_t max_height = FLAGS_max_height;
  neb::block_height_t height = FLAGS_height;

  // std::string cur_path = neb::fs::cur_dir();
  // std::string db_path = neb::fs::join_path(cur_path, "test_data.db");
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readwrite);

  auto f_keys = [](rocksdb::Iterator *it) {
    for (it->SeekToFirst(); it->Valid(); it->Next()) {
      LOG(INFO) << it->key().ToString();
    }
  };

  auto f_raw_val = [&](rocksdb::Iterator *it) {};

  auto f_set_nbre_max_height = [&]() {
    rs.put("nbre_max_height",
           neb::util::number_to_byte<neb::util::bytes>(max_height));
  };
  f_set_nbre_max_height();

  auto f_lib_height = [&]() {
    auto height_bytes = rs.get("blockchain_lib");
    LOG(INFO) << neb::util::byte_to_number<neb::block_height_t>(height_bytes);
  };
  // f_lib_height();

  auto f_block_hash = [&](neb::block_height_t height) {
    auto block_hash_bytes =
        rs.get_bytes(neb::util::number_to_byte<neb::util::bytes>(height));
    LOG(INFO) << neb::util::byte_to_number<neb::block_height_t>(
        block_hash_bytes);
  };
  // f_block_hash(height);

  // rs.show_all(f_keys);

  return 0;
}
