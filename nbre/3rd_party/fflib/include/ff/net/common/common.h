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
#include <cstdint>
#include <memory>
#include <mutex>
#include <tuple>
#include <vector>

#include <boost/any.hpp>
#include <boost/asio.hpp>
#include <boost/noncopyable.hpp>
#include <cstdint>
#include <cstdio>
#include <fstream>
#include <functional>
#include <iostream>
#include <map>
#include <memory>
#include <queue>
#include <stddef.h>
#include <string>
#include <thread>
#include <type_traits>
#include <vector>

namespace ff {
namespace net {
typedef std::string String;
enum net_mode {
  real_net = 1,
  simu_net,
  single_net,
};

enum {
  protobuf_wrapper_pkg_type = 1,
  retrans_pkg_type = 10,

  simu_udp_open_type,
  simu_udp_send_pkg_type,
  simu_udp_close_type,

  ffnet_internal_reserve,
};
} // namespace net

} // namespace ff
