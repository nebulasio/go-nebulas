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
#include "ff/util/is_callable.h"
#include <functional>
#include <type_traits>

namespace ff {
namespace util {

namespace internal {
template <class F, class AT>
struct is_function_with_arg_type_impl : public std::false_type {};

template <class Ret, class AT>
struct is_function_with_arg_type_impl<Ret (*)(AT), AT> : public std::true_type {
};

template <class Ret, class AT>
struct is_function_with_arg_type_impl<Ret(AT), AT> : public std::true_type {};

template <class Ret, class AT, class C>
struct is_function_with_arg_type_impl<Ret (C::*)(AT) const, AT>
    : public std::true_type {};

template <class Ret, class AT, class C>
struct is_function_with_arg_type_impl<Ret (C::*)(AT), AT>
    : public std::true_type {};

template <class Ret>
struct is_function_with_arg_type_impl<Ret (*)(), void> : public std::true_type {
};

template <class Ret>
struct is_function_with_arg_type_impl<Ret(), void> : public std::true_type {};

template <class Ret, class C>
struct is_function_with_arg_type_impl<Ret (C::*)() const, void>
    : public std::true_type {};

template <class Ret, class C>
struct is_function_with_arg_type_impl<Ret (C::*)(), void>
    : public std::true_type {};

/////////////////////////
template <class F, class AT1, class AT2>
struct is_function_with_two_arg_type_impl : public std::false_type {};

template <class Ret, class AT1, class AT2>
struct is_function_with_two_arg_type_impl<Ret (*)(AT1, AT2), AT1, AT2>
    : public std::true_type {};

template <class Ret, class AT1, class AT2>
struct is_function_with_two_arg_type_impl<Ret(AT1, AT2), AT1, AT2>
    : public std::true_type {};

template <class Ret, class C, class AT1, class AT2>
struct is_function_with_two_arg_type_impl<Ret (C::*)(AT1, AT2), AT1, AT2>
    : public std::true_type {};

template <class Ret, class C, class AT1, class AT2>
struct is_function_with_two_arg_type_impl<Ret (C::*)(AT1, AT2) const, AT1, AT2>
    : public std::true_type {};

template <class Ret, class AT1>
struct is_function_with_two_arg_type_impl<Ret (*)(AT1), AT1, void>
    : public std::true_type {};

template <class Ret, class AT1>
struct is_function_with_two_arg_type_impl<Ret(AT1), AT1, void>
    : public std::true_type {};

template <class Ret, class C, class AT1>
struct is_function_with_two_arg_type_impl<Ret (C::*)(AT1), AT1, void>
    : public std::true_type {};

template <class Ret, class C, class AT1>
struct is_function_with_two_arg_type_impl<Ret (C::*)(AT1) const, AT1, void>
    : public std::true_type {};

template <class Ret, class AT1>
struct is_function_with_two_arg_type_impl<Ret (*)(AT1), void, AT1>
    : public std::true_type {};

template <class Ret, class AT1>
struct is_function_with_two_arg_type_impl<Ret(AT1), void, AT1>
    : public std::true_type {};

template <class Ret, class C, class AT1>
struct is_function_with_two_arg_type_impl<Ret (C::*)(AT1), void, AT1>
    : public std::true_type {};

template <class Ret, class C, class AT1>
struct is_function_with_two_arg_type_impl<Ret (C::*)(AT1) const, void, AT1>
    : public std::true_type {};

template <class Ret>
struct is_function_with_two_arg_type_impl<Ret (*)(), void, void>
    : public std::true_type {};

template <class Ret>
struct is_function_with_two_arg_type_impl<Ret(), void, void>
    : public std::true_type {};

template <class Ret, class C>
struct is_function_with_two_arg_type_impl<Ret (C::*)(), void, void>
    : public std::true_type {};

template <class Ret, class C>
struct is_function_with_two_arg_type_impl<Ret (C::*)() const, void, void>
    : public std::true_type {};
} // namespace internal

template <class F, class AT, bool flag>
struct is_functor_with_arg_type : std::false_type {};

template <class F, class AT> struct is_functor_with_arg_type<F, AT, true> {
  typedef decltype(&std::remove_reference<F>::type::operator()) FT;
  const static bool value =
      internal::is_function_with_arg_type_impl<FT, AT>::value;
};

template <class F, class AT> struct is_function_with_arg_type {
  typedef typename std::remove_reference<F>::type FT;
  const static bool s_is_class = std::is_class<FT>::value;
  const static bool s_is_bind_expr = std::is_bind_expression<FT>::value;
  const static bool value = s_is_bind_expr || is_functor_with_arg_type < FT, AT,
                    s_is_class && !s_is_bind_expr &&
                            is_callable<FT>::value > ::value ||
                        internal::is_function_with_arg_type_impl<FT, AT>::value;
};
////////////////////////////////

template <class F, class AT1, class AT2, bool flag>
struct is_functor_with_two_arg_type : public std::false_type {};

template <class F, class AT1, class AT2>
struct is_functor_with_two_arg_type<F, AT1, AT2, true> {
  typedef decltype(&std::remove_reference<F>::type::operator()) FT;
  const static bool value =
      internal::is_function_with_two_arg_type_impl<FT, AT1, AT2>::value;
};

template <class F, class AT1, class AT2>
struct is_function_with_two_arg_type
    : public std::conditional<
          std::is_class<typename std::remove_reference<F>::type>::value,
          is_functor_with_two_arg_type<
              F, AT1, AT2,
              std::is_class<typename std::remove_reference<F>::type>::value &&
                  is_callable<F>::value>,
          internal::is_function_with_two_arg_type_impl<
              typename std::remove_reference<F>::type, AT1, AT2>>::type {};

} // end namespace util
}  // end namespace ff

