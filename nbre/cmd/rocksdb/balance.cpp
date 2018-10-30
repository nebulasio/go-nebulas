
#include "fs/nbre_storage.h"
#include "fs/proto/block.pb.h"
#include "fs/util.h"

#include <boost/program_options.hpp>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {

  po::options_description desc("Rocksdb read account balance");
  desc.add_options()("help", "show help message")(
      "db_path", po::value<std::string>(), "Database file directory");

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

  std::string db_path = vm["db_path"].as<std::string>();
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readonly);

  neb::block_height_t height = 1;
  auto block = std::make_shared<corepb::Block>();
  auto height_hash =
      rs.get_bytes(neb::util::number_to_byte<neb::util::bytes>(height));
  auto block_bytes = rs.get_bytes(height_hash);
  bool ret = block->ParseFromArray(block_bytes.value(), block_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse block failed");
  }
  LOG(INFO) << block->height();
  for (auto &tx : block->transactions()) {
    std::string from = tx.from();
    auto bytes_from = neb::util::string_to_byte(tx.from());
    LOG(INFO) << bytes_from.to_base58();
    auto ret = rs.get(bytes_from.to_hex());
  }

  rs.close_database();

  return 0;
}
