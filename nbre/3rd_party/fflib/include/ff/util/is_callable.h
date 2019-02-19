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
#include <type_traits>

namespace ff {
namespace util {

namespace internal {
template <typename T>
struct is_callable_object {
 private:
  typedef char(&yes)[1];
  typedef char(&no)[2];

  struct Fallback {
    void operator()();
  };
  struct Derived : T, Fallback {};

  template <typename U, U>
  struct Check;

  template <typename>
  static yes test(...);

  template <typename C>
  static no test(Check<void (Fallback::*)(), &C::operator()>*);

 public:
  static const bool value = sizeof(test<Derived>(0)) == sizeof(yes);
};

template <typename T>
struct is_callable_base : public std::false_type {};

template <class Ret, class C, class... Args>
struct is_callable_base<Ret (C::*)(Args...) const> : public std::true_type {};

template <class Ret, class C, class... Args>
struct is_callable_base<Ret (C::*)(Args...)> : public std::true_type {};

template <class Ret, class... Args>
struct is_callable_base<Ret(Args...)> : public std::true_type {};

template <class Ret, class... Args>
struct is_callable_base<Ret (*)(Args...)> : public std::true_type {};

} // namespace internal

template <class T>
struct is_callable
    : std::conditional<
          std::is_class<typename std::remove_reference<T>::type>::value,
          internal::is_callable_object<typename std::remove_reference<T>::type>,
          internal::is_callable_base<typename std::remove_reference<T>::type>>::
          type {};
}  // end namespace utils
}  // end namespace ff

