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
#ifndef FF_PARA_PARA_WAIT_TRAITS_H_
#define FF_PARA_PARA_WAIT_TRAITS_H_
#include "ff/functionflow/common/common.h"
namespace ff {
template <class RT>
class para;
namespace internal {
template <class T1, class T2>
class wait_and;
template <class T1, class T2>
class wait_or;

class wait_all;
class wait_any;
}  // end namespace internal

template <class T>
struct is_para_or_wait : public std::false_type {};

template <class T>
struct is_para_or_wait<para<T> > : public std::true_type {};

template <class T1, class T2>
struct is_para_or_wait<internal::wait_and<T1, T2> > : public std::true_type {};

template <class T1, class T2>
struct is_para_or_wait<internal::wait_or<T1, T2> > : public std::true_type {};

template <>
struct is_para_or_wait<internal::wait_all> : public std::true_type {};

template <>
struct is_para_or_wait<internal::wait_any> : public std::true_type {};
}  // end namespace ff

#endif
