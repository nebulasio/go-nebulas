// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//

#include "common/util/byte.h"
#include "common/util/version.h"
#include "fs/blockchain.h"
#include "fs/proto/ir.pb.h"
#include "fs/util.h"
#include <boost/foreach.hpp>
#include <boost/format.hpp>
#include <boost/program_options.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <fstream>
#include <iostream>

namespace po = boost::program_options;

void read_block_json_file(const std::string &conf_fp,
                          boost::property_tree::ptree &json_root) {
  try {
    boost::property_tree::read_json(conf_fp, json_root);
  } catch (const boost::property_tree::ptree_error &e) {
    throw e.what();
  }
}

std::string get_db_path_for_read() {
  std::string cur_path = neb::fs::cur_dir();
  std::string db_path = neb::fs::join_path(cur_path, "test/data/read-data.db/");
  return db_path;
}

std::string get_db_path_for_write() {
  std::string cur_path = neb::fs::cur_dir();
  std::string db_path =
      neb::fs::join_path(cur_path, "test/data/write-data.db/");
  return db_path;
}

void ir_block_gen(const std::string &conf_fp, const neb::util::bytes &bytes) {

  boost::property_tree::ptree pt;
  read_block_json_file(conf_fp, pt);

  std::string db_path = get_db_path_for_read();
  std::shared_ptr<neb::fs::blockchain> blockchain_ptr =
      std::make_shared<neb::fs::blockchain>(db_path);

  std::shared_ptr<corepb::Block> lib_block_ptr =
      blockchain_ptr->load_LIB_block();
  neb::block_height_t lib_height = lib_block_ptr->height();
  std::cout << "block height " << lib_height + 1 << std::endl;

  corepb::Block block;
  boost::property_tree::ptree transactions = pt.get_child("transactions");
  BOOST_FOREACH (boost::property_tree::ptree::value_type &v, transactions) {
    boost::property_tree::ptree tx = v.second;

    corepb::Transaction *transaction = block.add_transactions();
    transaction->set_hash(tx.get<std::string>("hash"));
    transaction->set_from(tx.get<std::string>("from"));
    transaction->set_to(tx.get<std::string>("to"));
    transaction->set_value(tx.get<std::string>("value"));
    transaction->set_nonce(tx.get<int32_t>("nonce"));
    transaction->set_timestamp(tx.get<int32_t>("timestamp"));

    corepb::Data *data = transaction->mutable_data();
    data->set_type(tx.get_child("data").get<std::string>("type"));
    data->set_payload(neb::util::byte_to_string(bytes));

    transaction->set_chain_id(tx.get<int32_t>("chain_id"));
    transaction->set_gas_price(tx.get<std::string>("gas_price"));
    transaction->set_gas_limit(tx.get<std::string>("gas_limit"));
  }

  block.set_height(lib_height + 1);

  std::string serial_block;
  bool ret = block.SerializeToString(&serial_block);
  if (!ret) {
    throw std::runtime_error("serialize block proto failed");
  }

  blockchain_ptr.reset();

  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readwrite);

  neb::util::bytes height_bytes =
      neb::util::number_to_byte<neb::util::bytes>(lib_height + 1);
  neb::util::bytes block_hash_bytes =
      neb::util::string_to_byte(height_bytes.to_hex());
  rs.put_bytes(height_bytes, block_hash_bytes);
  rs.put_bytes(block_hash_bytes, neb::util::string_to_byte(serial_block));
  rs.put("blockchain_lib", block_hash_bytes);
}

int main(int argc, char *argv[]) {

  po::options_description desc("Generate Block");
  desc.add_options()("help", "show help message")(
      "ir_binary", po::value<std::string>(), "IR binary proto file")(
      "block_conf", po::value<std::string>(), "Block info configuration file");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);
  if (vm.count("help")) {
    std::cout << desc << "\n";
    return 1;
  }

  if (!vm.count("ir_binary")) {
    std::cout << "You must specify \"ir_binary\"!" << std::endl;
    return 1;
  }
  if (!vm.count("block_conf")) {
    std::cout << "You must specify \"block_conf\"!" << std::endl;
    return 1;
  }
  std::ifstream ifs;

  try {
    std::string ir_fp = vm["ir_binary"].as<std::string>();
    ifs.open(ir_fp, std::ios::in | std::ios::binary);

    if (!ifs.is_open()) {
      throw std::invalid_argument(
          boost::str(boost::format("can't open file %1%") % ir_fp));
    }

    ifs.seekg(0, ifs.end);
    std::ifstream::pos_type size = ifs.tellg();
    if (size > 128 * 1024) {
      throw std::invalid_argument("IR file too large!");
    }

    neb::util::bytes buf(size);

    ifs.seekg(0, ifs.beg);
    ifs.read((char *)buf.value(), buf.size());
    if (!ifs)
      throw std::invalid_argument(boost::str(
          boost::format("Read IR file error: only %1% could be read") %
          ifs.gcount()));

    std::string conf_fp = vm["block_conf"].as<std::string>();
    ir_block_gen(conf_fp, buf);

  } catch (std::exception &e) {
    ifs.close();
    std::cout << e.what() << std::endl;
  }

  return 0;
}
