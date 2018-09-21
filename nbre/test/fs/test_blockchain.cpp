#include "fs/blockchain.h"
#include "common/util/base58.h"
#include "common/util/base64.h"
#include "fs/proto/block.pb.h"
#include "fs/rocksdb_storage.h"
#include "fs/util.h"

void get_tail_block_from_rocksdb() {
  std::string cur_path = neb::fs::cur_dir();
  std::string db_path = neb::fs::join_path(cur_path, "test/data/data.db/");
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readonly);
  auto tail_block_hash = rs.get(neb::fs::blockchain::Block_LIB);
  auto tail_bytes = rs.get_bytes(tail_block_hash);

  corepb::Block block;
  block.ParseFromArray(tail_bytes.value(), tail_bytes.size());
  LOG(INFO) << block.height();
  rs.close_database();
}

int main(int argc, char *argv[]) {

  std::string db_path = "/home/chmwang/data.db";

  std::shared_ptr<neb::fs::blockchain> blockchain_ptr =
      std::make_shared<neb::fs::blockchain>(db_path);
  std::shared_ptr<corepb::Block> block_ptr =
      blockchain_ptr->load_block_with_height(1);

  auto header = block_ptr->header();
  LOG(INFO) << "timestamp: " << header.timestamp();

  auto txs = block_ptr->transactions();
  auto tx = txs.begin();
  LOG(INFO) << tx->nonce() << ',' << tx->chain_id();

  std::string hash = tx->hash();
  neb::util::bytes hash_b = neb::util::string_to_byte(hash);
  std::string hash_hex = hash_b.to_hex();
  LOG(INFO) << hash_hex;

  std::string from = tx->from();
  neb::util::bytes from_b = neb::util::string_to_byte(from);
  std::string from_base58 = from_b.to_base58();
  LOG(INFO) << from_base58;

  std::string value = tx->value();
  neb::util::bytes value_b = neb::util::string_to_byte(value);
  std::string value_hex = value_b.to_hex();
  LOG(INFO) << value.size() << ',' << value_hex.size() << ',' << value_hex;

  std::string price = tx->gas_price();
  neb::util::bytes price_b = neb::util::string_to_byte(price);
  std::string price_hex = price_b.to_hex();
  LOG(INFO) << price.size() << ',' << price_hex.size() << ',' << price_hex;

  auto data = tx->data();
  std::string type = data.type();
  LOG(INFO) << type;

  std::string payload = data.payload();
  neb::util::bytes payload_b = neb::util::string_to_byte(payload);
  std::string payload_base64 = payload_b.to_base64();
  LOG(INFO) << payload_base64;

  // std::string s_raw = "xx";
  // std::string s_encode = neb::encode_base64(s_raw);
  // std::string s_decode;
  // neb::decode_base64(s_encode, s_decode);
  // LOG(INFO) << s_raw << ',' << s_decode << ',' << s_encode;

  neb::util::bytes b_decode = neb::util::bytes::from_base64("eHg=");
  LOG(INFO) << b_decode.to_hex();
  for (size_t i = 0; i < 10; i++) {
    std::string b_encode = b_decode.to_base64();
    b_decode = neb::util::bytes::from_base64(b_encode);
    LOG(INFO) << b_encode;
  }
  // LOG(INFO) << b_decode;

  return 0;
}
