// Copyright (C) 2018-2019 go-nebulas authors
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

#pragma once

#include <vector>
#include <mutex>
#include <unordered_map>

template<class Key, class Value, class Lock = std::mutex, uint32_t MaxLength = 128>
class LRU_MAP{
  public:
    typedef std::unordered_map<Key, Value> mp_t;
    typedef std::vector<Key> vc_t;

    LRU_MAP(){
      m_mp = std::unique_ptr<mp_t>(new mp_t());
      m_vec = std::unique_ptr<vc_t>(new vc_t());
    }
    ~LRU_MAP(){}

    void set(Key key, Value value){
      if(m_vec->size() >= MaxLength){
        Key old_key = *(m_vec->begin());
        m_vec->erase(m_vec->begin());
        m_mp->erase(old_key);
      }
      m_vec->push_back(key);
      m_mp->insert(std::make_pair(key, value));
    }

    bool find(Key key){
      auto target = m_mp->find(key);
      if(target != m_mp->end())
        return true;
      return false;
    }

    // Caller needs to make sure the key exists, call find firstly
    Value get(Key key){
      auto target = m_mp->find(key);
      return target->second;
    }

    uint32_t size(){
      return m_mp->size();
    }

  private:
    std::unique_ptr<vc_t> m_vec;
    std::unique_ptr<mp_t> m_mp;
    //Lock m_lock;              Thread safe will be crucial if parallelism is triggered, for now, we do not use it for performance's sake
};