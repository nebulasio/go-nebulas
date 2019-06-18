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
#include "fs/ir_manager/ir_processor.h"
#include "common/configuration.h"
#include "compatible/compatible_check_interface.h"
#include "core/execution_context.h"
#include "core/net_ipc/nipc_pkg.h"
#include "fs/blockchain.h"
#include "fs/ir_manager/api/ir_list.h"
#include "fs/storage.h"
#include "jit/cpp_ir.h"
#include "runtime/auth/auth_handler.h"
#include "runtime/auth/auth_table.h"
#include "util/npr.h"
#include "util/persistent_flag.h"
#include "util/persistent_type.h"
#include "util/wakeable_queue.h"

namespace neb {
namespace fs {
ir_processor::ir_processor(class storage *s, class blockchain *bc)
    : m_storage(s), m_blockchain(bc) {
  m_ir_list = std::make_unique<ir_list>(s);
  m_auth_handler = std::make_unique<rt::auth::auth_handler>(m_ir_list.get());
  m_failed_flag = std::make_unique<util::persistent_flag>(
      s, configuration::instance().nbre_failed_flag_name());
  m_nbre_block_height = std::make_unique<util::persistent_type<block_height_t>>(
      s, configuration::instance().nbre_max_height_name());
}

optional<nbre::NBREIR>
ir_processor::get_ir_with_version(const std::string &name, version_t v) {
  try {
    return m_ir_list->get_ir(name, v);
  } catch (...) {
    return none;
  }
}

optional<bytes>
ir_processor::get_ir_brief_key_with_height(const std::string &name,
                                           block_height_t h) {
  try {
    return m_ir_list->get_ir_brief_key_with_height(name, h);
  } catch (...) {
    return none;
  }
}

optional<nbre::NBREIR> ir_processor::get_ir_with_height(const std::string &name,
                                                        block_height_t h) {
  try {
    return m_ir_list->find_ir_at_height(name, h);
  } catch (...) {
    return none;
  }
}

std::vector<nbre::NBREIR> ir_processor::get_ir_depends(const nbre::NBREIR &ir) {
  std::vector<nbre::NBREIRDepend> depends;
  std::vector<nbre::NBREIR> irs;
  std::unordered_set<bytes> visited;

  auto to_bytes = [](const std::string &str, version_t v) {
    return number_to_byte<bytes>(v) + str;
  };

  irs.push_back(ir);
  for (int i = 0; i < ir.depends_size(); ++i) {
    depends.push_back(ir.depends(i));
  }

  while (!depends.empty()) {
    nbre::NBREIRDepend d = depends.back();
    depends.pop_back();
    auto key = to_bytes(d.name(), d.version());
    if (visited.find(key) != visited.end()) {
      continue;
    }
    visited.insert(key);
    optional<nbre::NBREIR> dir = get_ir_with_version(d.name(), d.version());
    if (dir == none) {
      LOG(WARNING) << "cannot get ir with name: " << d.name() << ", "
                   << d.version();
      return std::vector<nbre::NBREIR>();
    }
    if (dir->ir_type() != ir_type::llvm) {
      return std::vector<nbre::NBREIR>();
    }
    irs.push_back(*dir);
    for (int i = 0; i < dir->depends_size(); ++i) {
      depends.push_back(dir->depends(i));
    }
  }
  return irs;
}

std::vector<std::string> ir_processor::get_ir_names() const {
  return m_ir_list->get_ir_names();
}

std::vector<version_t>
ir_processor::get_ir_versions(const std::string &name) const {
  return m_ir_list->get_ir_versions(name);
}

void ir_processor::parse_irs(
    util::wakeable_queue<std::shared_ptr<nbre_ir_transactions_req>> &q_txs) {
  block_height_t last_height = m_nbre_block_height->get();
  while (!q_txs.empty()) {
    auto ele = q_txs.try_pop_front();
    if (!ele.first) {
      break;
    }
    auto h = ele.second->get<p_height>();
    if (h < last_height + 1) {
      continue;
    }
    if (h > last_height + 1) {
      parse_missed_blocks_between(last_height + 1, h);
    }
    const auto &txs = ele.second->get<p_ir_transactions>();
    parse_block_with_height_and_txs(h, txs);
    last_height = m_nbre_block_height->get();
  }
}

void ir_processor::parse_missed_blocks_between(block_height_t start_height,
                                               block_height_t end_height) {

  std::string ir_tx_type = neb::configuration::instance().ir_tx_payload_type();

  for (block_height_t h = start_height; h < end_height; h++) {
    auto block = m_blockchain->load_block_with_height(h);
    std::vector<corepb::Transaction> txs;

    for (auto &tx : block->transactions()) {
      auto &data = tx.data();
      const std::string &type = data.type();
      if (type == ir_tx_type) {
        txs.push_back(tx);
      }
    }
    parse_block_with_height_and_txs(h, txs);
  }
}
void ir_processor::parse_block_with_height_and_txs(
    block_height_t height, const std::vector<std::string> &str_txs) {
  std::vector<corepb::Transaction> txs;
  if (str_txs.empty()) {
    parse_block_with_height_and_txs(height, txs);
    return;
  }
  for (auto &tx_str : str_txs) {
    corepb::Transaction tx;
    bool ret = tx.ParseFromString(tx_str);
    if (!ret) {
      continue;
    }
    if (!util::is_npr_tx(tx)) {
      continue;
    }
    txs.push_back(tx);
  }
  parse_block_with_height_and_txs(height, txs);
}

void ir_processor::parse_block_with_height_and_txs(
    block_height_t height, const std::vector<corepb::Transaction> &txs) {
  if (!m_failed_flag->test()) {
    m_failed_flag->set();
    parse_ir_txs(height, txs);
  }
  m_nbre_block_height->set(height);
  m_failed_flag->clear();
}

void ir_processor::parse_ir_txs(block_height_t height,
                                const std::vector<corepb::Transaction> &txs) {
  for (auto &tx : txs) {
    if (!util::is_npr_tx(tx)) {
      continue;
    }
    try {
      nbre::NBREIR npr = util::extract_npr(tx);

        if (npr.height() <= height) {
          LOG(WARNING) << "npr available height is less than tx height, "
                          "ignore it, tx:"
                       << string_to_byte(tx.hash()).to_base58();
          continue;
        }

        auto &name = npr.name();
        auto version = npr.version();

        if (m_ir_list->ir_exist(name, version)) {
          LOG(INFO) << "ignore ir name: " << name
                    << " with version: " << version;
          continue;
        }
        address_t from = to_address(tx.from());
        if (util::is_auth_npr(from, name)) {

          //! The available height is too small, we need ignore this.
          auto compiled_ir = compile_payload_code(npr);
          m_ir_list->write_ir(npr, compiled_ir);
          // m_auth_handler->handle_auth_npr(compiled_ir);
          LOG(INFO) << "got auth table npr";
          continue;
      }

      block_height_t height = npr.height();
      if (!m_auth_handler->is_ir_legitimate(name, from, height)) {
        continue;
      }

      auto compiled_ir = compile_payload_code(npr);
      m_ir_list->write_ir(npr, compiled_ir);
    } catch (const std::exception &e) {
      LOG(ERROR) << "got exception " << e.what()
                 << " for tx: " << string_to_byte(tx.hash()).to_base58();
      continue;
    }
  }
}

nbre::NBREIR ir_processor::compile_payload_code(const nbre::NBREIR &raw_ir) {
  nbre::NBREIR nbre_ir = raw_ir;
  nbre_ir.set_ir_type(::neb::ir_type::invalid);
  nbre_ir.set_ir(std::string());

  if (raw_ir.ir_type() == ::neb::ir_type::cpp) {
    std::stringstream ss;
    ss << raw_ir.name();
    ss << raw_ir.version();
    //! For compatible reason, we may ignore this.
    compatible::compatible_check_interface *cc =
        core::context->compatible_checker();
    bool need_compile = cc->is_ir_need_compile(raw_ir.name(), raw_ir.version());
    if (!need_compile) {
      return nbre_ir;
    }

    cpp::cpp_ir ci(std::make_pair(ss.str(), raw_ir.ir()));
    neb::bytes ir = ci.llvm_ir_content();
    nbre_ir.set_ir(neb::byte_to_string(ir));
    nbre_ir.set_ir_type(::neb::ir_type::llvm);
  }
  return nbre_ir;
}

} // namespace fs
} // namespace neb
