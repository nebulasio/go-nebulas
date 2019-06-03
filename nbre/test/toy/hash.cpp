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

#include <functional>
#include <iostream>
#include <set>
#include <string>
#include <unordered_set>
#include <vector>

struct elf_hash_32bit {
  std::size_t operator()(const std::string &key) const {
    std::size_t hash = 0U;
    const std::size_t mask = 0xF0000000;
    for (std::string::size_type i = 0; i < key.length(); ++i) {
      hash = (hash << 4U) + key[i];
      std::size_t x = hash & mask;
      if (x != 0)
        hash ^= (x >> 24);
      hash &= ~x;
    }
    std::cout << "for key " << key << " hash " << hash << std::endl;
    return hash;
  }
};

void show() {

  std::vector<std::string> v;
  v.push_back("n1YxyaBzoRogNh72eri7HBGijCAtcHpf9nm,");
  v.push_back("n1GV9UScgncwU6KKL9T18mCo2S6uAE69SWs,");
  v.push_back("n1X2SXyEKej7GZgAreXDCkiT59qaYKBDcYi,");

  // std::unordered_set<std::string> s;
  std::unordered_set<std::string, elf_hash_32bit> s;
  for (size_t i = 0; i < v.size(); i++) {
    s.insert(v[i]);
  }

  for (auto it = s.begin(); it != s.end(); it++) {
    std::cout << *it << std::endl;
  }
}

int main() {
  show();
  return 0;
}
