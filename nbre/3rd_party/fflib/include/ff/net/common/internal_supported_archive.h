/***********************************************
The MIT License (MIT)

Copyright (c) 2018 Athrun Arthur <athrunarthur@gmail.com>

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

//! We use this file to support user-defined data types for rchive
#pragma once

#include "ff/net/common/common.h"
namespace ff {
namespace net {
namespace internal {
template <class T, bool ArithmeticFlag = std::is_arithmetic<T>::value>
class internal_supported_archive_type : public std::false_type {};
template <class T>
class internal_supported_archive_type<T, true> : public std::true_type {};

template <bool ArithmeticFlag>
class internal_supported_archive_type<std::string, ArithmeticFlag>
    : public std::true_type {};
template <class T, size_t N, bool ArithmeticFlag>
class internal_supported_archive_type<T[N], ArithmeticFlag>
    : public std::true_type {};

template <class T, bool ArithmeticFlag>
class internal_supported_archive_type<std::vector<T>, ArithmeticFlag>
    : public std::true_type {};

} // namespace internal

template <class T> class udt_marshaler {};

} // namespace net
} // namespace ff
