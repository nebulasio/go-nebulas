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
#include "ff/net/network/end_point.h"
#include "ff/net/network/tcp_connection_base.h"

namespace ff {
namespace net {

class tcp_server;

//////////////////////////////////////////////////////
class net_tcp_connection : public net_tcp_connection_base {
public:
  net_tcp_connection(io_service &ioservice, pkg_packer *bs, event_handler *eh,
                     const std::vector<tcp_pkg_handler *> &rph,
                     tcp_server *pSvr);

  inline tcp_server *get_tcp_server() { return m_pTCPServer; }

  virtual ~net_tcp_connection() {}

protected:
  friend class net_tcp_server;

  void start();

  tcp_server *m_pTCPServer;
};

// end class TCPConnection
// typedef boost::shared_ptr<tcp_connection> tcp_connection_ptr;

//////////////////////////////////////////////////////
class tcp_server : public boost::noncopyable {
public:
  tcp_server(io_service &ioservice, pkg_packer *bs, event_handler *eh,
             const std::vector<tcp_pkg_handler *> &rph, const tcp_endpoint &ep);

  virtual ~tcp_server() {}
  inline boost::asio::io_service &ioservice() { return m_oIOService; }

  virtual void start_accept() = 0;

  virtual void close() = 0;

protected:
  io_service &m_oIOService;
  tcp_endpoint m_oAcceptEP;
  pkg_packer *m_pBS;
  event_handler *m_pEH;
  std::vector<tcp_pkg_handler *> m_pRPH;
};
typedef std::shared_ptr<tcp_server> tcp_server_ptr;

class net_tcp_server : public tcp_server {
public:
  net_tcp_server(io_service &ioservice, pkg_packer *bs, event_handler *eh,
                 const std::vector<tcp_pkg_handler *> &rph,
                 const tcp_endpoint &ep);

  inline boost::asio::ip::tcp::acceptor &acceptor() { return m_oAcceptor; }

  virtual void start_accept();

  virtual void close();

protected:
  void handle_accept(tcp_connection_base_ptr pNewConn,
                     const boost::system::error_code &error);

protected:
  boost::asio::ip::tcp::acceptor m_oAcceptor;
};
} // namespace net
} // namespace ff

