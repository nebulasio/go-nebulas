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

//! This file is used to define compile-time error messages.
#ifndef FF_COMMON_ERROR_MSG_H_
#define FF_COMMON_ERROR_MSG_H_

namespace ff {
template <class F>
struct Please_Check_The_Assert_Msg {
  const static bool value = false;
};  // end struct Please_Check_The_Assert_Msg
}  // end namespace ff

#define FF_EM_CALL_THEN_WITHOUT_CALL_PAREN                                     \
  "You can only call *then* right after calling "                              \
  "operator(...), like this, \n\tpara<> a;\n\ta([](){...}).then(...);"

#define FF_EM_CALL_WITH_TYPE_MISMATCH                                          \
  "The function's return type in operator(...) doesn't match "                 \
  "para<...>"

#define FF_EM_THEN_WITH_TYPE_MISMATCH                                          \
  "The parameter's type in then's callback function doesn't "                  \
  "match associated para<...>'s return type"

#define FF_EM_THEN_WITH_NON_FUNC_TYPE                                          \
  "*then* must take a callable object as parameter, like "                     \
  "lambda, functor, or std::function"

#define FF_EM_CALL_SQPAREN_AFTER_PAREN                                         \
  "You can only call operator[] before calling calling "                       \
  "operator(...), instead of after calling operator(...)!"

#define FF_EM_WRONG_USE_SQPAREN                                                \
  "1mYou can only wait for a para<...> object or a dependency "                \
  "expression while using operator[]"

#define FF_EM_CALL_PAREN_AFTER_PAREN                                           \
  "You cannot call operator() after operator()"

#define FF_EM_CALL_PAREN_WITH_NO_FUNC                                          \
  "You should call operator() with a callable object, like a "                 \
  "function pointer, a functor, or a lambda"

#define FF_EM_CALL_PAREN_WITH_WRONG_PARAM                                      \
  "You pass operator() a function with wrong parameter "                       \
  "types"

#define FF_EM_CALL_PAREN_WITH_WRONG_RET                                        \
  "You pass operator() a function with wrong return "                          \
  "type"

#define FF_EM_COMBINE_PARA_AND_OTHER                                           \
  "Cannot combine para<...> object and other object as "                       \
  "dependency expression"

#define FF_EM_USE_PARACONTAINER_INSTEAD_OF_GROUP                               \
  "paragroup is only for data parallelism (for_each), use "                    \
  "paracontainer to hold multiple para<...> objects!"

#define FF_EM_THEN_FUNC_TYPE_MISMATCH                                          \
  "The function's type in *then* doesn't match with "                          \
  "associated dependency expression"

#define FF_EM_CALL_NO_SUPPORT_FOR_PARA                                         \
  "Currently, we don't support that returning a para object "                  \
  "in another para object"

#define FF_EM_CALL_FOR_EACH_WITHOUT_FUNCTION                                   \
  "The 3rd parameter's type should be callable, maybe a "                      \
  "function, or a functor."
#define FF_EM_CALL_FOR_EACH_WRONG_FUNCTION                                     \
  "The callback function of the 3rd parameter has wrong input "                \
  "type."

//! for utilities
#define FF_EM_CALL_FOR_EACH_WITH_NO_FUNC                                       \
  "You should call for_each with a callable "                                  \
  "function/functor"
#define FF_EM_CALL_FOR_EACH_WITH_WRONG_PARAM                                   \
  "The callback function for for_each should have T (from "                    \
  "thread_local_var<T>) as parameter."

#endif
