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

#include "fs/nbre_storage.h"
#include "common/util/byte.h"

namespace neb {
namespace fs {

nbre_storage::nbre_storage(const std::string &path,
                           const std::string &bc_path) {
  m_storage = std::unique_ptr<rocksdb_storage>(new rocksdb_storage());
  m_storage->open_database(path, storage_open_for_readwrite);

  m_blockchain = std::unique_ptr<blockchain>(new blockchain(bc_path));
}

std::shared_ptr<nbre::NBREIR>
nbre_storage::read_nbre_by_name_version(const std::string &name,
                                        uint64_t version) {
  std::shared_ptr<nbre::NBREIR> nbre_ir = std::make_shared<nbre::NBREIR>();
  if (!m_storage) {
    return nbre_ir;
  }

  std::string name_version = name + std::to_string(version);
  neb::util::bytes nbre_bytes = m_storage->get(name_version);
  nbre_ir->ParseFromArray(nbre_bytes.value(), nbre_bytes.size());
  return nbre_ir;
}

void nbre_storage::write_nbre_by_height(block_height_t height) {
  if (!m_storage || !m_blockchain) {
    return;
  }

  auto block = m_blockchain->load_block_with_height(height);

  for (auto &tx : block->transactions()) {
    auto &data = tx.data();
    const std::string &type = data.type();

    if (type.compare(m_payload_type) == 0) {
      const std::string &payload = data.payload();
      neb::util::bytes payload_bytes = neb::util::string_to_byte(payload);

      std::shared_ptr<nbre::NBREIR> nbre_ir = std::make_shared<nbre::NBREIR>();
      nbre_ir->ParseFromArray(payload_bytes.value(), payload_bytes.size());
      const std::string &name = nbre_ir->name();
      const uint64_t version = nbre_ir->version();
      std::string name_version = name + std::to_string(version);

      m_storage->put(name_version, payload_bytes);
    }
  }
}
} // namespace fs
} // namespace neb
