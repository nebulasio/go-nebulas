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
#pragma once
#include "common/common.h"

namespace neb {
class version {
public:
  inline version() { m_data.m_data = 0; }
  version(uint64_t data) { m_data.m_data = data; };
  version(uint32_t major_version, uint16_t minor_version,
          uint16_t patch_version) {
    m_data.m_detail.m_major_version = major_version;
    m_data.m_detail.m_minor_version = minor_version;
    m_data.m_detail.m_patch_version = patch_version;
  }

  version(const version &) = default;
  version &operator=(const version &) = default;

  inline uint32_t major_version() const {
    return m_data.m_detail.m_major_version;
  }
  inline uint16_t minor_version() const {
    return m_data.m_detail.m_minor_version;
  }
  inline uint16_t patch_version() const {
    return m_data.m_detail.m_patch_version;
  }
  inline uint32_t &major_version() { return m_data.m_detail.m_major_version; }
  inline uint16_t &minor_version() { return m_data.m_detail.m_minor_version; }
  inline uint16_t &patch_version() { return m_data.m_detail.m_patch_version; }

  inline uint64_t data() const { return m_data.m_data; }

  inline version &from_string(const std::string &str) {
    std::vector<int> vs;
    std::istringstream f(str);
    std::string s;
    while (std::getline(f, s, '.')) {
      vs.push_back(std::atoi(s.c_str()));
    }
    if (vs.size() != 3) {
      throw std::invalid_argument("invalid version string");
    }
    m_data.m_detail.m_major_version = vs[0];
    m_data.m_detail.m_minor_version = vs[1];
    m_data.m_detail.m_patch_version = vs[2];
    return *this;
  }

  friend inline std::ostream &operator<<(std::ostream &stream,
                                         const version &v) {
    stream << v.m_data.m_detail.m_major_version << ','
           << v.m_data.m_detail.m_minor_version << ','
           << v.m_data.m_detail.m_patch_version;
    return stream;
  }

  friend inline bool operator<(const version &v1, const version &v2) {
    if (v1.m_data.m_detail.m_major_version <
        v2.m_data.m_detail.m_major_version) {
      return true;
    } else if (v1.m_data.m_detail.m_minor_version <
               v2.m_data.m_detail.m_minor_version) {
      return true;
    } else if (v1.m_data.m_detail.m_patch_version <
               v2.m_data.m_detail.m_patch_version) {
      return true;
    }
    return false;
  }

  friend inline bool operator>(const version &v1, const version &v2) {
    return v2 < v1;
  }

  friend inline bool operator>=(const version &v1, const version &v2) {
    return !(v1 < v2);
  }

  friend inline bool operator<=(const version &v1, const version &v2) {
    return !(v1 > v2);
  }

  friend inline bool operator==(const version &v1, const version &v2) {
    return v1 >= v2 && v2 >= v1;
  }

  friend inline bool operator!=(const version &v1, const version &v2) {
    return !(v1 == v2);
  }


protected:
  union _version_data {
    uint64_t m_data;
    struct _version_detail {
      uint32_t m_major_version;
      uint16_t m_minor_version;
      uint16_t m_patch_version;
    };
    _version_detail m_detail;
  };

  _version_data m_data;
}; // end class version
}
