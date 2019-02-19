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
#include "ff/net/middleware/package.h"
#include "ff/net/middleware/pkg_handler.h"
#include "ff/net/network/asio_point.h"
#include "ff/net/network/end_point.h"
#include "ff/net/network/net_buffer.h"

namespace ff {
namespace net {
class udp_pkg_handler;

class udp_point : public asio_point {
public:
  udp_point(io_service &ioservice, pkg_packer *bs, event_handler *eh,
            const std::vector<udp_pkg_handler *> &rph, const udp_endpoint &ep);

  virtual ~udp_point();

  virtual void send(const package_ptr &pkg, const udp_endpoint &ep) = 0;

  virtual void close() = 0;

protected:
  udp_endpoint mo_self_endpoint;
  std::vector<udp_pkg_handler *> m_pRPH;
};

class net_udp_point : public udp_point {
public:
  net_udp_point(io_service &ioservice, pkg_packer *bs, event_handler *eh,
                const std::vector<udp_pkg_handler *> &rph,
                const udp_endpoint &ep);

  virtual ~net_udp_point();

  virtual void send(const package_ptr &pkg, const udp_endpoint &endpoint);

  virtual void close();

  void ticktack();

protected:
  void start_send();

  void handle_pkg_sent(const boost::system::error_code &ec,
                       std::size_t bytes_transferred);

  void actual_send_pkg(const package_ptr &pkg, const udp_endpoint &endpoint);

  void start_recv();

  void handle_received_pkg(const boost::system::error_code &error,
                           size_t bytes_transferred);

  void slice_and_dispatch_pkg(net_buffer *pBuf, const udp_endpoint &ep);

protected:
  typedef std::function<void()> func_t;
  typedef std::queue<func_t> tasks_t;
  typedef std::map<udp_endpoint, net_buffer *> recv_buffer_t;

  boost::asio::ip::udp::socket m_oSocket;
  udp_endpoint m_oRecvEndPoint;
  udp_endpoint m_oSendEndpoint;
  tasks_t m_oSendTasks;
  net_buffer m_oSendBuffer;
  net_buffer m_oTempBuffer;
  recv_buffer_t m_oRecvBuffer;
  std::map<uint32_t, udp_pkg_handler *> m_oRPHCache;
  bool m_bIsSending;
};

// end class UDPPoint
typedef std::shared_ptr<udp_point> udp_point_ptr;

} // namespace net
} // end namespace ff
