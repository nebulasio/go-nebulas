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
#ifdef PROTO_BUF_SUPPORT
#include <google/protobuf/descriptor.h>
#include <google/protobuf/message.h>
#endif

namespace ff {
namespace net {
class tcp_pkg_handler;

class tcp_connection_base : public asio_point {
public:
  tcp_connection_base(boost::asio::io_service &ioservice, pkg_packer *bs,
                      event_handler *eh,
                      const std::vector<tcp_pkg_handler *> &rph);

  virtual ~tcp_connection_base();

  virtual void send(const package_ptr &pkg) = 0;

#ifdef PROTO_BUF_SUPPORT
  virtual void send(const std::shared_ptr<::google::protobuf::Message> &pkg);
#endif

  virtual void close() = 0;

  inline const tcp_endpoint &remote_endpoint() const {
    return m_oRemoteEndpoint;
  }

protected:
  std::vector<tcp_pkg_handler *> m_pRPH;
  tcp_endpoint m_oRemoteEndpoint;
};

class net_tcp_connection_base
    : public tcp_connection_base,
      public std::enable_shared_from_this<net_tcp_connection_base> {
public:
  virtual ~net_tcp_connection_base();

  virtual void send(const package_ptr &pkg);

  virtual void close();

protected:
  net_tcp_connection_base(boost::asio::io_service &ioservice, pkg_packer *bs,
                          event_handler *eh,
                          const std::vector<tcp_pkg_handler *> &rph);

  void start_send();

  void start_recv();

  void handle_pkg_sent(boost::system::error_code ec,
                       std::size_t bytes_transferred);

  void handle_received_pkg(const boost::system::error_code &error,
                           size_t bytes_transferred);

  void slice_and_dispatch_pkg();

protected:
  friend class net_tcp_server;

  typedef std::queue<package_ptr> pkgs_t;
  boost::asio::ip::tcp::socket m_oSocket;
  net_buffer m_oRecvBuffer;
  net_buffer m_oSendBuffer;
  pkgs_t m_oToSendPkgs;
  bool m_bIsSending;
  std::map<uint32_t, tcp_pkg_handler *> m_oRPHCache;
};

typedef std::shared_ptr<tcp_connection_base> tcp_connection_base_ptr;

} // namespace net

} // namespace ff

