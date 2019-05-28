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

template<class Key, class Value, class Lock = std::mutex>
class LRU_MAP{
  public:
    typedef std::unordered_map<Key, Value> mp_t;
    typedef std::vector<Key> vc_t;

    LRU_MAP(uint32_t buffer_size):m_buffer_size(buffer_size){
      m_mp = std::unique_ptr<mp_t>(new mp_t());
      m_vec = std::unique_ptr<vc_t>(new vc_t());
    }
    LRU_MAP(){
      m_mp = std::unique_ptr<mp_t>(new mp_t());
      m_vec = std::unique_ptr<vc_t>(new vc_t());
    }
    ~LRU_MAP(){}

    void set(Key key, Value value){
      std::cout<<"%%%%%%%%%% LRU vec size before: "<<m_vec->size()<<std::endl;
      if(m_vec->size() >= m_buffer_size){
        std::cout<<"%%%%%%%%%%%% The LRU_MAP vec size is: "<<m_vec->size()<<std::endl;
        Key old_key = *(m_vec->begin());
        m_vec->erase(m_vec->begin());
        m_mp->erase(old_key);
        std::cout<<"%%%%%%%%%%% Erased an old key"<<std::endl;
      }
      m_vec->push_back(key);
      m_mp->insert(std::make_pair(key, value));
      std::cout<<"%%%%%%%%%% LRU vec size after: "<<m_vec->size()<<std::endl;
    }

    bool find(const Key& key){
      auto target = m_mp->find(key);
      if(target != m_mp->end())
        return true;
      return false;
    }

    // Caller needs to make sure the key exists, call find firstly
    Value get(const Key& key){
      auto target = m_mp->find(key);
      return target->second;
    }

    uint32_t size(){
      return m_mp->size();
    }

    void clear(){
      if(m_vec != nullptr)
        m_vec->clear();
      if(m_mp != nullptr)
        m_mp->clear();
    }

  private:
    std::unique_ptr<vc_t> m_vec;
    std::unique_ptr<mp_t> m_mp;
    uint32_t m_buffer_size=128;
    //Lock m_lock;              Thread safe will be crucial if parallelism is triggered, for now, we do not use it for performance's sake
};