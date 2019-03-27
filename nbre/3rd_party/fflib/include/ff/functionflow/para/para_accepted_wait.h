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
template <class PT, class WT>
class para_accepted_wait {
  para_accepted_wait &operator=(const para_accepted_wait<PT, WT> &) = delete;

 public:
  typedef typename PT::ret_type ret_type;
  typedef typename std::remove_reference<WT>::type WT_t;
  typedef typename WT_t::ret_type wret_type;

  para_accepted_wait(const para_accepted_wait<PT, WT> &) = default;
  para_accepted_wait(PT &p, const WT &w) : m_refP(p), m_oWaiting(w) {}

  template <class F> // for f with valid parameter
  auto operator()(F &&f) -> typename std::enable_if<
      is_wait_compatible_with_func<WT_t, F>::value &&
          //! utils::function_args_traits<F>::is_no_args &&
          std::is_same<ret_type,
                       typename util::function_res_traits<F>::ret_type>::value,
      ::ff::internal::para_accepted_call<PT, ret_type>>::type {
    auto pp =
        std::make_shared<::ff::internal::para_impl_wait<ret_type, WT_t, F>>(
            m_oWaiting, std::forward<F>(f));
    ::ff::internal::para_impl_base_ptr<ret_type> pTask =
        std::dynamic_pointer_cast<::ff::internal::para_impl_base<ret_type>>(pp);
    m_refP.m_pImpl = pTask;
    ::ff::internal::schedule(pTask);
    return ::ff::internal::para_accepted_call<PT, ret_type>(m_refP);
  }

  template <class F> // for f is not a function
  auto operator()(F &&f) -> typename std::enable_if<
      !util::is_callable<F>::value,
      ::ff::internal::para_accepted_call<PT, ret_type>>::type {
    static_assert(Please_Check_The_Assert_Msg<F>::value,
                  FF_EM_CALL_PAREN_WITH_NO_FUNC);
  }

  template <class F> // for f with invalid params
  auto operator()(F &&f) -> typename std::enable_if<
      util::is_callable<F>::value &&
          std::is_same<ret_type, typename util::function_res_traits<
                                     F>::ret_type>::value &&
          !is_wait_compatible_with_func<WT_t, F>::value,
      ::ff::internal::para_accepted_call<PT, ret_type>>::type {
    static_assert(Please_Check_The_Assert_Msg<F>::value,
                  FF_EM_CALL_PAREN_WITH_WRONG_PARAM);
  }

  template <class F> // for f with invalid ret
  auto operator()(F &&f) -> typename std::enable_if<
      util::is_callable<F>::value &&
          !std::is_same<ret_type,
                        typename util::function_res_traits<F>::ret_type>::value,
      ::ff::internal::para_accepted_call<PT, ret_type>>::type {
    static_assert(Please_Check_The_Assert_Msg<F>::value,
                  FF_EM_CALL_PAREN_WITH_WRONG_RET);
  }

  template <class F>
  void then(F &&f) {
    static_assert(Please_Check_The_Assert_Msg<F>::value,
                  FF_EM_CALL_THEN_WITHOUT_CALL_PAREN);
  }

 protected:
  PT &m_refP;
  WT m_oWaiting;
};  // end class para_accepted_wait;
