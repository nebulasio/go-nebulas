
#include "common/util/version.h"
#include "fs/nbre_storage.h"
#include "fs/proto/block.pb.h"
#include "fs/proto/ir.pb.h"
#include "fs/rocksdb_storage.h"
#include "fs/util.h"

int main(int argc, char *argv[]) {

  std::string cur_path = neb::fs::cur_dir();
  std::string db_path = neb::fs::join_path(cur_path, "test_data.db");
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readonly);

  auto f_keys = [](rocksdb::Iterator *it) {
    for (it->SeekToFirst(); it->Valid(); it->Next()) {
      LOG(INFO) << it->key().ToString();
    }
  };
  rs.show_all(f_keys);

  return 0;
}
