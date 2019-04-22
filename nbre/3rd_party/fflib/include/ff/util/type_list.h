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

#include <tuple>
namespace ff {
namespace util {
template <typename... ARGS> struct type_list {};
template <typename A1, typename... ARGS> struct type_list<A1, ARGS...> {
  typedef A1 first_type;
  typedef type_list<ARGS...> sub_type_list;
  const static int len = 1 + type_list<ARGS...>::len;
};

template <> struct type_list<> { const static int len = 0; };

template <typename T1, typename T2> struct is_same_type {
  const static bool value = false;
};

template <typename T1> struct is_same_type<T1, T1> {
  const static bool value = true;
};

template <int V> struct int_number_type { const static int value = V; };

template <typename T, typename TL> struct is_type_in_type_list {
  const static bool value =
      is_same_type<T, typename TL::first_type>::value ||
      is_type_in_type_list<T, typename TL::sub_type_list>::value;
};

template <typename T> struct is_type_in_type_list<T, type_list<>> {
  const static bool value = false;
};

template <typename TL1, typename TL2> struct is_contain_types {};

template <typename... ARGS1, typename T, typename... ARGS2>
struct is_contain_types<type_list<ARGS1...>, type_list<T, ARGS2...>> {
  const static bool value =
      is_type_in_type_list<T, type_list<ARGS1...>>::value &&
      is_contain_types<type_list<ARGS1...>, type_list<ARGS2...>>::value;
};

template <typename... ARGS1>
struct is_contain_types<type_list<ARGS1...>, type_list<>> {
  const static bool value = true;
};

template <typename T1, typename T2> struct merge_type_list {};

template <typename... ARGS1, typename... ARGS2>
struct merge_type_list<type_list<ARGS1...>, type_list<ARGS2...>> {
  typedef type_list<ARGS1..., ARGS2...> type;
};

template <typename T1, typename T2> struct merge_tuple {};

template <typename... ARGS1, typename... ARGS2>
struct merge_tuple<std::tuple<ARGS1...>, std::tuple<ARGS2...>> {
  typedef std::tuple<ARGS1..., ARGS2...> type;
};

template <typename TL> struct convert_type_list_to_tuple {};

template <typename T, typename T1, typename... TS>
struct convert_type_list_to_tuple<type_list<T, T1, TS...>> {
  typedef
      typename merge_tuple<std::tuple<T>, typename convert_type_list_to_tuple<
                                              type_list<T1, TS...>>::type>::type
          type;
};
template <typename T> struct convert_type_list_to_tuple<type_list<T>> {
  typedef std::tuple<T> type;
};
template <> struct convert_type_list_to_tuple<type_list<>> {
  typedef std::tuple<> type;
};
/////////////////
//
template <typename... ARGS>
struct check_if_replicate {}; // end struct check_if_replicate

template <typename T, typename... ARGS> struct check_if_replicate<T, ARGS...> {
  const static bool value =
      is_type_in_type_list<T, type_list<ARGS...>>::value ||
      check_if_replicate<ARGS...>::value;
};

template <typename T> struct check_if_replicate<T> {
  const static bool value = false;
};

template <typename TC, typename TL> struct get_index_of_type_in_typelist {};
template <typename TC, typename T1, typename... TS>
struct get_index_of_type_in_typelist<TC, type_list<T1, TS...>> {
  const static int value = std::conditional<
      is_same_type<TC, T1>::value, int_number_type<0>,
      int_number_type<1 + get_index_of_type_in_typelist<
                              TC, type_list<TS...>>::value>>::type::value;
};

template <typename TC> struct get_index_of_type_in_typelist<TC, type_list<>> {
  const static int value = -1;
};

template <typename TL, int index> struct get_type_at_index_in_typelist {
  const static int value = -1;
  typedef void type;
};

template <typename T, typename... TS, int index>
struct get_type_at_index_in_typelist<type_list<T, TS...>, index> {
  typedef
      typename std::conditional<index == 0, T,
                                typename get_type_at_index_in_typelist<
                                    type_list<TS...>, index - 1>::type>::type
          type;
};

template <typename TL> struct extract_content_type_list {};

template <typename T1, typename T2, typename... TS>
struct extract_content_type_list<type_list<T1, T2, TS...>> {
  typedef typename merge_type_list<
      type_list<typename T1::type>,
      typename extract_content_type_list<type_list<T2, TS...>>::type>::type
      type;
};

template <typename T1> struct extract_content_type_list<type_list<T1>> {
  typedef type_list<typename T1::type> type;
};

template <> struct extract_content_type_list<type_list<>> {
  typedef type_list<> type;
};
} // namespace util
} // namespace ff
