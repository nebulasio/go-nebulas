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
#include "ff/net/network/tcp_client.h"
#include "ff/net/common/defines.h"
#include "ff/net/middleware/event_handler.h"
#include "ff/net/network/events.h"

namespace ff {
namespace net {
using namespace ::ff::net::event;
using namespace ::ff::net::event::more;

net_tcp_client::net_tcp_client(io_service &ioservice, pkg_packer *bs,
                               event_handler *eh,
                               const std::vector<tcp_pkg_handler *> &rph,
                               const tcp_endpoint &ep)
    : net_tcp_connection_base(ioservice, bs, eh, rph) {
  boost::system::error_code ec;
  m_pEH->triger<tcp_client_start_connection>(ep);
  m_oSocket.async_connect(ep, [this](const boost::system::error_code &ec) {
    handle_connected(ec);
  });
}

void net_tcp_client::handle_connected(const boost::system::error_code &ec) {
  if (!ec) {
    // LOG(INFO) << "Get connection succ";
    m_iPointState = state_valid;
    m_oRemoteEndpoint = m_oSocket.remote_endpoint();
    m_pEH->triger<tcp_client_get_connection_succ>(this);
    start_recv();
  } else {
    m_iPointState = state_error;
    // LOG(WARNING) << "Get connection error!";
    m_pEH->triger<tcp_client_conn_error>(this, ec);
  }
}

} // namespace net
} // namespace ff
