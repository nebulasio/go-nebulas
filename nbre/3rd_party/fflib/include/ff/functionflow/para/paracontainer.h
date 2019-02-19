/***********************************************
  The MIT License (MIT)

  Copyright (c) 2012 Athrun Arthur <athrunarthur@gmail.com>

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in
  all copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
  THE SOFTWARE.
 *************************************************/
#ifndef FF_PARA_PARA_CONTAINER_H_
#define FF_PARA_PARA_CONTAINER_H_
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/para/para.h"
#include "ff/functionflow/para/para_helper.h"
#include "ff/functionflow/para/paras_with_lock.h"
#include "ff/functionflow/runtime/rtcmn.h"
#include <cmath>

namespace ff {

namespace internal {
class wait_all;
class wait_any;

}  // end namespace internal

class paracontainer {
 public:
  typedef void ret_type;

 public:
   paracontainer() : m_pEntities(new ::ff::internal::paras_with_lock()) {}

   para<void> &operator[](int index) {
     std::lock_guard<ff::spinlock> _l(m_pEntities->lock);
     return (*m_pEntities).entities[index];
  }
  size_t size() const {
    std::lock_guard<ff::spinlock> _l(m_pEntities->lock);
    return m_pEntities->entities.size();
  }
  ~paracontainer() {}

  template <typename Func_t>
  void add(const Func_t& f) {
    para<void> p;
    p(f);
    add(p);
  }

  void add(const para<void>& p) {
    std::lock_guard<ff::spinlock> _l(m_pEntities->lock);
    m_pEntities->entities.push_back(p);
  }

  void clear() {
    std::lock_guard<ff::spinlock> _l(m_pEntities->lock);
    m_pEntities->entities.clear();
  }

 protected:
   typedef std::shared_ptr<::ff::internal::paras_with_lock> Entities_t;

   friend ::ff::internal::wait_all all(paracontainer &pg);
   friend ::ff::internal::wait_any any(paracontainer &pg);
   std::shared_ptr<::ff::internal::paras_with_lock> &all_entities() {
     return m_pEntities;
  };

  std::shared_ptr<::ff::internal::paras_with_lock> m_pEntities;
};  // end class paracontainer

}  // end namespace ff

#endif
