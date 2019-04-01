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
#include "ff/net/common/common.h"
#include "ff/net/common/shared_buffer.h"
#include "ff/net/middleware/package.h"
#include "ff/net/network/end_point.h"

namespace ff {
namespace net {
class udp_point;
class tcp_connection_base;

class udp_pkg_handler {
public:
  virtual ~udp_pkg_handler() {}
  virtual void handle_pkg(udp_point *pPoint, const shared_buffer &buf,
                          const udp_endpoint &from) = 0;
  virtual bool is_pkg_to_handle(uint32_t pkg_id) = 0;
};

class tcp_pkg_handler {
public:
  virtual ~tcp_pkg_handler() {}
  virtual void handle_pkg(tcp_connection_base *pFrom,
                          const shared_buffer &buf) = 0;
  virtual bool is_pkg_to_handle(uint32_t pkg_id) = 0;
};

template <typename Base_, typename Ty_> class pkg_recv_callback {
public:
  typedef std::function<void(std::shared_ptr<Ty_>)> pkg_recv_handler_t;

  static void recv_handler(std::shared_ptr<Base_> pPkg,
                           const pkg_recv_handler_t &handler) {
    std::shared_ptr<Ty_> pConcretPkg =
        std::dynamic_pointer_cast<Ty_, Base_>(pPkg);
    handler(pConcretPkg);
  }
};

template <class Base_, class Ty_> class tcp_recv_callback {
public:
  typedef std::function<void(std::shared_ptr<Ty_>, tcp_connection_base *)>
      pkg_recv_handler_t;

  static void recv_handler(std::shared_ptr<Base_> pPkg,
                           tcp_connection_base *pConn,
                           const pkg_recv_handler_t &handler) {
    std::shared_ptr<Ty_> pConcretPkg =
        std::dynamic_pointer_cast<Ty_, Base_>(pPkg);
    handler(pConcretPkg, pConn);
  }
};

template <class Base_, class Ty_> class udp_recv_callback {
public:
  typedef std::function<void(std::shared_ptr<Ty_>, udp_point *, udp_endpoint)>
      pkg_recv_handler_t;

  static void recv_handler(std::shared_ptr<Base_> pPkg, udp_point *from,
                           udp_endpoint ep, const pkg_recv_handler_t &handler) {
    std::shared_ptr<Ty_> pConcretPkg =
        std::dynamic_pointer_cast<Ty_, Base_>(pPkg);
    handler(pConcretPkg, from, ep);
  }
};

template <class Ty_> class package_new_wrapper {
public:
  static package_ptr New() { return package_ptr(new Ty_()); }
};

} // namespace net
} // namespace ff
