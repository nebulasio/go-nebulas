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

#include "ff/functionflow/common/common.h"
#include "ff/functionflow/para/para.h"
#include "ff/functionflow/para/paracontainer.h"
#include "ff/functionflow/para/paragroup.h"
#include "ff/functionflow/para/wait.h"
#include "ff/functionflow/utilities/accumulator.h"
#include "ff/functionflow/utilities/hazard_pointer.h"
#include "ff/functionflow/utilities/miso_queue.h"
#include "ff/functionflow/utilities/scope_guard.h"
#include "ff/functionflow/utilities/simo_queue.h"
#include "ff/functionflow/utilities/single_assign.h"
#include "ff/functionflow/utilities/spin_lock.h"
#include "ff/functionflow/utilities/thread_local_var.h"

namespace ff {

template <class W>
void ff_wait(W&& wexpr) {
  (wexpr).then([]() {});
}  // end wait
template <class RT>
void ff_wait(para<RT>& sexpr) {
  ff_wait(sexpr && sexpr);
}
}  // end namespace ff

