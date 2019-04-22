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
#pragma once
#include "ff/util/internal/user_new_type.h"
#include "ff/util/preprocessor.h"
#include "ff/util/tuple_type.h"
#include "ff/util/type_list.h"
#include <memory>

namespace ff {
namespace util {

/*
 * ntobject is used to define a struct without writng your own class.
 * To use it, you need to declare a bunch of types first, like this
 *
 * define_nt(email, std::string, "email");
 * define_nt(uid, uint64_t, "uid");
 * define_nt(_name, std::string, "name");
 *
 * typedef ntobject<email, uid, name> myobject_t;
 *
 * Now you have your own structure -- myobject_t, you can set and get field in
 * it like
 *
 * myobject_t obj;
 * obj.set<_name>("xuepeng");
 * obj.set<email>("xp@example.com");
 * obj.set<uid>(128);
 *
 * std::string val_name = obj.get<name>();
 *
 * Further more, you can set multiple values with any order if you want,
 *
 * obj.set<name, email>("xuepeng", "xp@example.com");
 * obj.set<email, uid, name> ("xp@example.com", 128, "xuepeng");
 *
 *
 * */
template <typename... ARGS> class ntobject {
public:
  typedef typename util::type_list<ARGS...> type_list;

  ntobject() : m_content(new content_type()) {}

  ntobject<ARGS...> make_copy() const {
    ntobject<ARGS...> rt;
    *rt.m_content = *m_content;
    return rt;
  }

  template <typename CT>
  void set(const typename internal::nt_traits<CT>::type &val) {
    static_assert(is_type_in_type_list<CT, util::type_list<ARGS...>>::value,
                  "Cannot set a value that's not in the ntobject/row!");
    const static int index =
        get_index_of_type_in_typelist<CT, util::type_list<ARGS...>>::value;
    std::get<index>(*m_content) = val;
  }

  template <typename CT, typename CT1, typename... CARGS, typename... PARGS>
  void set(const typename internal::nt_traits<CT>::type &val,
           const typename internal::nt_traits<CT1>::type &val1,
           PARGS... params) {
    static_assert(is_type_in_type_list<CT, util::type_list<ARGS...>>::value,
                  "Cannot set a value that's not in the row!");
    static_assert(is_type_in_type_list<CT1, util::type_list<ARGS...>>::value,
                  "Cannot set a value that's not in the row!");
    const static int index =
        get_index_of_type_in_typelist<CT, util::type_list<ARGS...>>::value;
    std::get<index>(*m_content) = val;

    set<CT1, CARGS...>(val1, params...);
  }

  template <typename CT> typename internal::nt_traits<CT>::type get() const {
    static_assert(is_type_in_type_list<CT, util::type_list<ARGS...>>::value,
                  "Cannot get a value that's not in the ntobject/row!");
    const static int index =
        get_index_of_type_in_typelist<CT, util::type_list<ARGS...>>::value;
    return std::get<index>(*m_content);
  }

protected:
  typedef
      typename convert_type_list_to_tuple<typename nt_extract_content_type_list<
          util::type_list<ARGS...>>::type>::type content_type;
  std::unique_ptr<content_type> m_content;
};

template <typename... ARGS> class ntarray {
public:
  typedef ntobject<ARGS...> row_type;

  void push_back(row_type &&row) { m_collection.push_back(std::move(row)); }

  void clear() { m_collection.clear(); }

  size_t size() const { return m_collection.size(); }

  bool empty() const { return m_collection.empty(); }

  row_type &operator[](size_t index) { return m_collection[index]; }

  const row_type &operator[](size_t index) const { return m_collection[index]; }

protected:
  std::vector<row_type> m_collection;
};

template <typename T> struct is_ntobject { const static bool value = false; };

template <typename... ARGS> struct is_ntobject<ntobject<ARGS...>> {
  const static bool value = true;
};

template <typename T> struct is_ntarray { const static bool value = false; };

template <typename... ARGS> struct is_ntarray<ntarray<ARGS...>> {
  const static bool value = true;
};

} // namespace util
} // namespace ff

