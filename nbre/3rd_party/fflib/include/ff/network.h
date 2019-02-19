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
#include "ff/net/common/archive.h"
#include "ff/net/common/defines.h"
#include "ff/net/common/mout.h"
#include "ff/net/common/shared_buffer.h"
//#include "util/blocking_queue.h"
#include <boost/bind.hpp>

#include "ff/net/middleware/net_nervure.h"
#include "ff/net/middleware/typed_pkg_hub.h"

#include "ff/net/framework/application.h"
#include "ff/net/framework/routine.h"
#include "ff/net/middleware/ntpackage.h"
#include "ff/net/network/events.h"

#ifdef PROTO_BUF_SUPPORT
#include "ff/net/extension/protobuf/proto_buf_wrapper_pkg.h"
#include "ff/net/extension/protobuf/protobuf_pkg_hub.h"
#endif

