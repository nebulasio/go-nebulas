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
#include "ff/net/network/tcp_connection_base.h"
#include "ff/net/common/defines.h"
#include "ff/net/middleware/event_handler.h"
#include "ff/net/middleware/pkg_handler.h"
#include "ff/net/middleware/pkg_packer.h"
#include "ff/net/network/events.h"

#ifdef PROTO_BUF_SUPPORT
#include "ff/net/extension/protobuf/proto_buf_wrapper_pkg.h"
#endif
namespace ff {
namespace net {
using namespace ::ff::net::event;
using namespace ::ff::net::event::more;

tcp_connection_base::tcp_connection_base(
    io_service &ioservice, pkg_packer *bs, event_handler *eh,
    const std::vector<tcp_pkg_handler *> &rph)
    : asio_point(ioservice, bs, eh), m_pRPH(rph) {}

tcp_connection_base::~tcp_connection_base() {
  // LOG(INFO) << "~tcp_connection_base()";
}

#ifdef PROTO_BUF_SUPPORT
void tcp_connection_base::send(
    const std::shared_ptr<::google::protobuf::Message> &pkg) {
  std::shared_ptr<protobuf_wrapper_pkg> pMsg(new protobuf_wrapper_pkg(pkg));
  send(pMsg);
}
#endif

net_tcp_connection_base::net_tcp_connection_base(
    io_service &ioservice, pkg_packer *bs, event_handler *eh,
    const std::vector<tcp_pkg_handler *> &rph)
    : tcp_connection_base(ioservice, bs, eh, rph), m_oSocket(ioservice),
      m_bIsSending(false) {}

net_tcp_connection_base::~net_tcp_connection_base() {
  if (m_iPointState == state_valid)
    close();
}

void net_tcp_connection_base::send(const package_ptr &pkg) {
  m_oToSendPkgs.push(pkg);
  if (!m_bIsSending) {
    m_oIOService.post([this]() { start_send(); });
  }
  // m_oIOService.post([this, pkg]() {
  // if (m_iPointState != state_valid) {
  // m_pEH->triger<tcp_pkg_send_failed>(this, pkg);
  // return;
  //}
  // m_oToSendPkgs.push(pkg);
  // if (!m_bIsSending) {
  // m_bIsSending = true;
  // start_send();
  //}
  //});
}

void net_tcp_connection_base::start_send() {
  while (m_oSendBuffer.filled() < 64 * 1024) {
    package_ptr pPkg;
    if (m_oToSendPkgs.pop(pPkg)) {
      m_pPacker->pack(m_oSendBuffer, pPkg);
    } else {
      break;
    }
  }
  if (m_oSendBuffer.length() != 0) {
    m_pEH->triger<tcp_start_send_stream>(
        this, boost::asio::buffer_cast<const char *>(m_oSendBuffer.readable()),
        boost::asio::buffer_size(m_oSendBuffer.readable()));

    m_oSocket.async_write_some(
        boost::asio::buffer(m_oSendBuffer.readable()),
        [this](boost::system::error_code ec, std::size_t bt) {
          handle_pkg_sent(ec, bt);
        });
    m_bIsSending = true;
  }
}

void net_tcp_connection_base::handle_pkg_sent(boost::system::error_code ec,
                                              std::size_t bytes_transferred) {
  m_bIsSending = false;
  if (!ec) {
    m_pEH->triger<tcp_send_stream_succ>(this, bytes_transferred);
    m_oSendBuffer.erase_buffer(bytes_transferred);
    // LOG(INFO) << "pkg sent " << bytes_transferred << " bytes, to "
    //<< m_oRemoteEndpoint.address().to_string();
    if (m_oSendBuffer.length() != 0) {
      m_pEH->triger<tcp_start_send_stream>(
          this,
          boost::asio::buffer_cast<const char *>(m_oSendBuffer.readable()),
          boost::asio::buffer_size(m_oSendBuffer.readable()));

      m_oSocket.async_write_some(
          boost::asio::buffer(m_oSendBuffer.readable()),
          [this](boost::system::error_code ec, std::size_t bt) {
            handle_pkg_sent(ec, bt);
          });
      m_bIsSending = true;
    } else {
      start_send();
    }
  } else {
    m_iPointState = state_error;
    // LOG(WARNING) << "handle_pkg_sent, get error " << ec.message();
    m_pEH->triger<tcp_send_stream_error>(this, ec);
  }
}

void net_tcp_connection_base::start_recv() {
  try {
    m_pEH->triger<tcp_start_recv_stream>(m_oSocket.local_endpoint(),
                                         m_oSocket.remote_endpoint());
    // LOG(INFO) << "start_recv() on " <<
    // m_oRemoteEndpoint.address().to_string();
    m_oSocket.async_read_some(
        boost::asio::buffer(m_oRecvBuffer.writeable()),
        [this](boost::system::error_code ec, std::size_t bt) {
          handle_received_pkg(ec, bt);
        });
  } catch (std::system_error se) {
    // LOG(WARNING) << "start_recv(), remote_endpoint is disconnected!";
  }
}

void net_tcp_connection_base::handle_received_pkg(
    const boost::system::error_code &error, size_t bytes_transferred) {
  if (!error) {
    m_pEH->triger<tcp_recv_stream_succ>(this, bytes_transferred);
    m_oRecvBuffer.filled() += bytes_transferred;
    // LOG(INFO) << "recv pkg: " << bytes_transferred << " bytes, from "
    //<< m_oRemoteEndpoint.address().to_string();
    slice_and_dispatch_pkg();
    start_recv();
  } else {
    m_iPointState = state_error;
    // LOG(WARNING) << "handle_received_pkg(), Get error " << error.message()
    //<< " from " << m_oRemoteEndpoint.address().to_string();
    m_pEH->triger<tcp_recv_stream_error>(this, error);
  }
}

void net_tcp_connection_base::slice_and_dispatch_pkg() {
  std::list<shared_buffer> sbs = m_pPacker->split(m_oRecvBuffer);
  for (std::list<shared_buffer>::iterator it = sbs.begin(); it != sbs.end();
       ++it) {
    // LOG(INFO) << "slice_and_dispatch_pkg";
    shared_buffer sb = *it;
    const char *pBuf = sb.buffer();
    uint32_t pkg_id;
    ff::net::deseralize(pBuf, pkg_id);
    bool got_pkg_handler = false;
    std::map<uint32_t, tcp_pkg_handler *>::iterator cit =
        m_oRPHCache.find(pkg_id);
    if (cit != m_oRPHCache.end()) {
      got_pkg_handler = true;
      tcp_pkg_handler *ph = cit->second;
      ph->handle_pkg(this, sb);
    } else {
      for (size_t i = 0; i < m_pRPH.size(); ++i) {
        if (m_pRPH[i]->is_pkg_to_handle(pkg_id)) {
          m_oRPHCache.insert(std::make_pair(pkg_id, m_pRPH[i]));
          m_pRPH[i]->handle_pkg(this, sb);
          got_pkg_handler = true;
        }
      }
    }

    if (!got_pkg_handler) {
      // LOG(WARNING)
      //<< "slice_and_dispatch_pkg(), cannot find handler for pkg id: "
      //<< pkg_id;
    }
  }
}

void net_tcp_connection_base::close() {
  m_oSocket.close();
  m_iPointState = state_closed;
}

} // namespace net

} // namespace ff
