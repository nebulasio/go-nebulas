
#include "common/util/version.h"
#include "fs/nbre_storage.h"
#include "fs/proto/block.pb.h"
#include "fs/proto/ir.pb.h"
#include "fs/rocksdb_storage.h"
#include "fs/util.h"
#include <boost/program_options.hpp>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {

  po::options_description desc("Rocksdb read and write");
  desc.add_options()("help", "show help message")(
      "db_path", po::value<std::string>(),
      "Database file directory")("max_height", po::value<neb::block_height_t>(),
                                 "Nbre max height setting");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);

  if (vm.count("help")) {
    std::cout << desc << "\n";
    return 1;
  }

  if (!vm.count("db_path")) {
    std::cout << "You must specify \"db_path\"!" << std::endl;
    return 1;
  }
  if (!vm.count("max_height")) {
    std::cout << "You must specify \"max_height\"!" << std::endl;
    return 1;
  }

  std::string db_path = vm["db_path"].as<std::string>();
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readonly);

  auto f_keys = [](rocksdb::Iterator *it) {
    for (it->SeekToFirst(); it->Valid(); it->Next()) {
      LOG(INFO) << it->key().ToString();
    }
  };

  auto f_raw_val = [&](rocksdb::Iterator *it) {};

  neb::block_height_t max_height = vm["max_height"].as<neb::block_height_t>();
  auto f_set_nbre_max_height = [&]() {
    rs.put("nbre_max_height",
           neb::util::number_to_byte<neb::util::bytes>(max_height));
  };
  // f_set_nbre_max_height();

  auto f_lib_height = [&]() {
    auto block_hash_bytes = rs.get("blockchain_lib");
    auto block_bytes = rs.get_bytes(block_hash_bytes);

    std::shared_ptr<corepb::Block> block = std::make_shared<corepb::Block>();
    bool ret = block->ParseFromArray(block_bytes.value(), block_bytes.size());
    LOG(INFO) << "blockchain lib height: " << block->height();
  };
  f_lib_height();

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
