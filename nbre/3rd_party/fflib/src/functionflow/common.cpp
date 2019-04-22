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
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/para/exception.h"

namespace ff {

exe_state exe_state_and(exe_state e1, exe_state e2) {
  if (e1 == exe_state::exe_empty || e2 == exe_state::exe_empty)
    throw empty_para_exception();
  if (e1 == e2)
    return e1;
  return exe_state::exe_wait;
}

exe_state exe_state_or(exe_state e1, exe_state e2) {
  if (e1 == exe_state::exe_empty || e2 == exe_state::exe_empty)
    throw empty_para_exception();
  if (e1 == exe_state::exe_over || e2 == exe_state::exe_over)
    return exe_state::exe_over;
  return exe_state::exe_wait;
}
} // end namespace ff
