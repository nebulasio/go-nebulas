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
template <typename... Types>
struct merge_tuples {
  typedef std::tuple<Types...> type;
};

template <>
struct merge_tuples<> {
  typedef std::tuple<> type;
};

template <typename Type>
struct merge_tuples<Type> {
  typedef std::tuple<Type> type;
};

template <typename... Types>
struct merge_tuples<std::tuple<Types...>> {
  typedef std::tuple<Types...> type;
};

template <typename... Types1, typename... Types2>
struct merge_tuples<std::tuple<Types1...>, std::tuple<Types2...>> {
  typedef std::tuple<Types1..., Types2...> type;
};

template <typename Type1, typename Type2>
struct merge_tuples<Type1, Type2> {
  typedef std::tuple<Type1, Type2> type;
};

template <typename Type, typename... Types>
struct merge_tuples<Type, std::tuple<Types...>> {
  typedef std::tuple<Type, Types...> type;
};

template <typename... Types, typename Type>
struct merge_tuples<std::tuple<Types...>, Type> {
  typedef std::tuple<Types..., Type> type;
};

template <typename... Types, typename Type, typename... Rest>
struct merge_tuples<std::tuple<Types...>, Type, Rest...> {
  typedef typename merge_tuples<Rest...>::type temp;
  typedef typename merge_tuples<std::tuple<Types..., Type>, temp>::type type;
};

template <typename Type, typename... Types, typename... Rest>
struct merge_tuples<Type, std::tuple<Types...>, Rest...> {
  typedef typename merge_tuples<Rest...>::type temp;
  typedef typename merge_tuples<std::tuple<Type, Types...>, temp>::type type;
};

template <typename... Types1, typename... Types2, typename... Rest>
struct merge_tuples<std::tuple<Types1...>, std::tuple<Types2...>, Rest...> {
  typedef typename merge_tuples<Rest...>::type temp;
  typedef
      typename merge_tuples<std::tuple<Types1..., Types2...>, temp>::type type;
};

template <typename Type1, typename Type2, typename... Rest>
struct merge_tuples<Type1, Type2, Rest...> {
  typedef typename merge_tuples<Rest...>::type temp;
  typedef typename merge_tuples<std::tuple<Type1, Type2>, temp>::type type;
};

} // end namespace util
}  // end namespace ff
