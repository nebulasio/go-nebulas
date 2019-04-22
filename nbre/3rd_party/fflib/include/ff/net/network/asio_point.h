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
#include "ff/net/network/net_buffer.h"

namespace ff {
namespace net {
class pkg_packer;
class event_handler;
using boost::asio::io_service;

class asio_point : public boost::noncopyable {
public:
  enum point_state {
    state_init,
    state_valid,
    state_closed,
    state_error,
  };

  asio_point(io_service &ioservice, pkg_packer *bs, event_handler *eh);
  virtual ~asio_point();

  inline pkg_packer *get_pkg_packer() const { return m_pPacker; }
  inline io_service &ioservice() { return m_oIOService; }
  inline const io_service &ioservice() const { return m_oIOService; }
  inline event_handler *get_event_handler() const { return m_pEH; }

protected:
  io_service &m_oIOService;
  pkg_packer *m_pPacker;
  event_handler *m_pEH;
  point_state m_iPointState;
};

typedef std::shared_ptr<asio_point> asio_point_ptr;

} // namespace net
} // namespace ff
