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
#ifndef FF_RUNTIME_TASK_QUEUE_H_
#define FF_RUNTIME_TASK_QUEUE_H_

#include "ff/functionflow/runtime/taskbase.h"

#ifndef USING_WORK_STEALING_QUEUE
#define USING_WORK_STEALING_QUEUE
#endif

#ifdef USING_MUTEX_STEAL_QUEUE
#undef USING_WORK_STEALING_QUEUE
#include "ff/runtime/mutex_steal_queue.h"
#endif

#ifdef USING_SPIN_STEAL_QUEUE
#undef USING_WORK_STEALING_QUEUE
#include "ff/runtime/spin_steal_queue.h"
#endif

#ifdef USING_GCC_WORK_STEALING_QUEUE
#undef USING_WORK_STEALING_QUEUE
#include "ff/runtime/gtwsq_fixed.h"
#endif


#ifdef USING_WORK_STEALING_QUEUE
#include "ff/functionflow/runtime/work_stealing_queue.h"
#endif

namespace ff {
namespace rt {

#ifdef USING_MUTEX_STEAL_QUEUE
typedef mutex_stealing_queue<task_base_ptr> work_stealing_queue;
#endif

#ifdef USING_SPIN_STEAL_QUEUE
typedef spin_stealing_queue<task_base_ptr, 8> work_stealing_queue;
#endif

#ifdef USING_GCC_WORK_STEALING_QUEUE
typedef gcc_work_stealing_queue<task_base_ptr, 8> work_stealing_queue;
#endif

#ifdef USING_WORK_STEALING_QUEUE
typedef default_work_stealing_queue<task_base_ptr, 8> work_stealing_queue;
#endif

}  // end namespace rt
}  // end namespace ff
#endif
