#include "fs/nbre_storage.h"
#include "fs/proto/block.pb.h"
#include "fs/proto/ir.pb.h"
#include "fs/rocksdb_storage.h"
#include "fs/util.h"

void set_block_ir_payload() {

  nbre::NBREIR nbre_ir;
  nbre_ir.set_name("xxx");
  nbre_ir.set_version(666);
  nbre_ir.set_height(456);
  nbre::NBREIRDepend *nbre_dep = nbre_ir.add_depends();
  nbre_dep->set_name("xix");
  nbre_dep->set_version(789);
  nbre_ir.set_ir("heh");

  std::string serial_payload;
  nbre_ir.SerializeToString(&serial_payload);

  corepb::Block block;
  corepb::Transaction *transaction = block.add_transactions();
  transaction->set_hash("transaction_hash");
  transaction->set_from("source");
  transaction->set_to("target");
  transaction->set_value("1");
  transaction->set_nonce(2);
  transaction->set_timestamp(3);

  corepb::Data *data = transaction->mutable_data();
  data->set_type("protocol");
  data->set_payload(serial_payload);

  transaction->set_chain_id(4);
  transaction->set_gas_price("5");
  transaction->set_gas_limit("6");

  std::string serial_block;
  block.SerializeToString(&serial_block);
  LOG(INFO) << block.ByteSizeLong();

  std::string cur_path = neb::fs::cur_dir();
  std::string db_path = neb::fs::join_path(cur_path, "test_data.db");
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readwrite);
  rs.put_bytes(neb::util::number_to_byte<neb::util::bytes>(
                   static_cast<neb::block_height_t>(1000)),
               neb::util::string_to_byte("block_hash"));
  rs.put("block_hash", neb::util::string_to_byte(serial_block));
}

void get_block_by_rocksdb() {

  std::string cur_path = neb::fs::cur_dir();
  std::string db_path = neb::fs::join_path(cur_path, "test_data.db");
  LOG(INFO) << db_path;
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readonly);

  auto val_bytes = rs.get_bytes(neb::util::number_to_byte<neb::util::bytes>(
      static_cast<neb::block_height_t>(456)));
  std::string val_str = neb::util::byte_to_string(val_bytes);
  LOG(INFO) << val_str;

  neb::util::bytes block_bytes = rs.get(val_str);
  std::shared_ptr<corepb::Block> block = std::make_shared<corepb::Block>();
  block->ParseFromArray(block_bytes.value(), block_bytes.size());
  for (auto &tx : block->transactions()) {
    LOG(INFO) << tx.hash();
    LOG(INFO) << tx.from();
  }
}

void get_block_by_blockchain() {

  std::string path = "/home/chmwang/go-nebulas/nbre/test_data.db";

  std::shared_ptr<neb::fs::blockchain> bc_ptr =
      std::make_shared<neb::fs::blockchain>(path);
  std::shared_ptr<corepb::Block> block_ptr =
      bc_ptr->load_block_with_height(456);
  for (auto &tx : block_ptr->transactions()) {
    LOG(INFO) << tx.hash();
    LOG(INFO) << tx.from();
    LOG(INFO) << tx.to();

    const std::string &payload = tx.data().payload();
    neb::util::bytes payload_bytes = neb::util::string_to_byte(payload);
    nbre::NBREIR nbre_ir;
    nbre_ir.ParseFromArray(payload_bytes.value(), payload_bytes.size());

    LOG(INFO) << nbre_ir.name();
    LOG(INFO) << nbre_ir.version();
    LOG(INFO) << nbre_ir.height();
    for (auto &dep : nbre_ir.depends()) {
      LOG(INFO) << dep.name();
      LOG(INFO) << dep.version();
    }
    LOG(INFO) << nbre_ir.ir();
  }
}

void nbre_storage_rw() {

  std::string path = "/home/chmwang/go-nebulas/nbre/test_data.db";

  std::shared_ptr<neb::fs::nbre_storage> nbre_ptr =
      std::make_shared<neb::fs::nbre_storage>(path, path);
  // nbre_ptr->write_nbre_by_height(1000);
  std::shared_ptr<nbre::NBREIR> nbre_ir_ptr =
      nbre_ptr->read_nbre_by_name_version("xxx", 666);

  nbre::NBREIR nbre_ir = *nbre_ir_ptr;
  LOG(INFO) << nbre_ir.name();
  LOG(INFO) << nbre_ir.version();
  LOG(INFO) << nbre_ir.height();
  for (auto &dep : nbre_ir.depends()) {
    LOG(INFO) << dep.name();
    LOG(INFO) << dep.version();
  }
  LOG(INFO) << nbre_ir.ir();
}

int main(int argc, char *argv[]) { return 0; }
