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
#include "ff/net/network/tcp_server.h"
#include "ff/net/common/defines.h"
#include "ff/net/middleware/event_handler.h"
#include "ff/net/network/events.h"

namespace ff {
namespace net {
using namespace ::ff::net::event;
using namespace ::ff::net::event::more;

net_tcp_connection::net_tcp_connection(
    io_service &ioservice, pkg_packer *bs, event_handler *eh,
    const std::vector<tcp_pkg_handler *> &rph, tcp_server *pSvr)
    : net_tcp_connection_base(ioservice, bs, eh, rph), m_pTCPServer(pSvr) {
  m_iPointState = state_valid;
}

void net_tcp_connection::start() {
  m_oRemoteEndpoint = m_oSocket.remote_endpoint();
  start_recv();
}

/////////
tcp_server::tcp_server(io_service &ioservice, pkg_packer *bs, event_handler *eh,
                       const std::vector<tcp_pkg_handler *> &rph,
                       const tcp_endpoint &ep)
    : m_oIOService(ioservice), m_oAcceptEP(ep), m_pBS(bs), m_pEH(eh),
      m_pRPH(rph) {}

net_tcp_server::net_tcp_server(io_service &ioservice, pkg_packer *bs,
                               event_handler *eh,
                               const std::vector<tcp_pkg_handler *> &rph,
                               const tcp_endpoint &ep)
    : tcp_server(ioservice, bs, eh, rph, ep), m_oAcceptor(ioservice, ep) {}

void net_tcp_server::start_accept() {
  auto pn = std::make_shared<net_tcp_connection>(m_oAcceptor.get_io_service(),
                                                 m_pBS, m_pEH, m_pRPH, this);
  tcp_connection_base_ptr pNewConn =
      std::dynamic_pointer_cast<tcp_connection_base>(pn);
  net_tcp_connection *ntc = static_cast<net_tcp_connection *>(pNewConn.get());

  // LOG(INFO) << "start_accept";
  m_pEH->triger<tcp_server_start_listen>(m_oAcceptor.local_endpoint());

  m_oAcceptor.async_accept(
      ntc->m_oSocket, [this, pNewConn](const boost::system::error_code &ec) {
        handle_accept(pNewConn, ec);
      });
}

void net_tcp_server::handle_accept(tcp_connection_base_ptr pNewConn,
                                   const boost::system::error_code &error) {
  if (!error) {
    net_tcp_connection *ntc = (net_tcp_connection *)pNewConn.get();
    // LOG(INFO) << "accept connection";
    ntc->start();
    m_pEH->triger<tcp_server_accept_connection>(pNewConn);
    start_accept();
  } else {
    m_pEH->triger<tcp_server_accept_error>(m_oAcceptEP, error);
  }
}

void net_tcp_server::close() {
  // m_oAcceptor.cancel();
  m_oAcceptor.close();
}

} // namespace net

} // namespace ff
