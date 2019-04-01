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
#include "ff/net/extension/protobuf/protobuf_pkg_hub.h"
#include "ff/net/common/defines.h"
#ifdef PROTO_BUF_SUPPORT

namespace ff {
namespace net {
protobuf_pkg_hub::pb_tcp_pkg_handler::pb_tcp_pkg_handler(
    protobuf_pkg_hub *p_hub)
    : m_pHub(p_hub) {}
protobuf_pkg_hub::pb_tcp_pkg_handler::~pb_tcp_pkg_handler() {}

void protobuf_pkg_hub::pb_tcp_pkg_handler::handle_pkg(
    tcp_connection_base *p_from, const shared_buffer &buf) {
  m_pHub->handle_tcp_pkg(p_from, buf);
}

bool protobuf_pkg_hub::pb_tcp_pkg_handler::is_pkg_to_handle(uint32_t pkg_id) {
  return pkg_id == protobuf_wrapper_pkg_type;
}

protobuf_pkg_hub::pb_udp_pkg_handler::pb_udp_pkg_handler(
    protobuf_pkg_hub *p_hub)
    : m_pHub(p_hub) {}
protobuf_pkg_hub::pb_udp_pkg_handler::~pb_udp_pkg_handler() {}

void protobuf_pkg_hub::pb_udp_pkg_handler::handle_pkg(udp_point *p_point,
                                                      const shared_buffer &buf,
                                                      const udp_endpoint &ep) {
  m_pHub->handle_udp_pkg(p_point, buf, ep);
}

bool protobuf_pkg_hub::pb_udp_pkg_handler::is_pkg_to_handle(uint32_t pkg_id) {
  return pkg_id == protobuf_wrapper_pkg_type;
}

protobuf_pkg_hub::protobuf_pkg_hub() {
  m_pTCPPkgHandler = new pb_tcp_pkg_handler(this);
  m_pUDPPkgHandler = new pb_udp_pkg_handler(this);
}

protobuf_pkg_hub::~protobuf_pkg_hub() {
  LOG(INFO) << "~protobuf_pkg_hub";
  delete m_pTCPPkgHandler;
  delete m_pUDPPkgHandler;
}

protobuf_pkg_hub::message_ptr
protobuf_pkg_hub::get_pkg(const shared_buffer &buf) {
  const char *pBuf = buf.buffer();
  uint32_t iTypeID;
  ffnet::deseralize(pBuf, iTypeID);
  if (iTypeID != protobuf_wrapper_pkg_type) {
    assert(0 && "can't find the type id for protobuf");
    return message_ptr();
  }

  protobuf_wrapper_pkg pkg;
  marshaler d(const_cast<const char *>(buf.buffer()), buf.length(),
              marshaler::deseralizer);
  pkg.arch(d);
  return pkg.protobuf_message();
}

void protobuf_pkg_hub::handle_tcp_pkg(tcp_connection_base *pFrom,
                                      const shared_buffer &buf) {
  message_ptr p_msg = get_pkg(buf);
  if (p_msg) {
    bool got_callback = false;
    std::string name(p_msg->GetDescriptor()->full_name());
    tcp_handlers_t::iterator it = m_oTCPHandlers.find(name);
    if (it != m_oTCPHandlers.end()) {
      got_callback = true;
      tcp_recv_handler_t handler = it->second;
      handler(p_msg, pFrom);
    }

    pkg_handlers_t::iterator tit = m_oPkgHandlers.find(name);
    if (tit != m_oPkgHandlers.end()) {
      got_callback = true;
      pkg_recv_handler_t handler = tit->second;
      handler(p_msg);
    }
    if (!got_callback) {
      // LOG_ERROR(frmwk)<<"protobuf_pkg::handle_tcp_pkg(), a message
      // with id: "<<name<<" has no handler!";
    }
  }
}

void protobuf_pkg_hub::handle_udp_pkg(udp_point *pPoint,
                                      const shared_buffer &buf,
                                      const udp_endpoint &from) {
  message_ptr p_msg = get_pkg(buf);
  if (p_msg) {
    bool got_callback = false;
    std::string name(p_msg->GetDescriptor()->full_name());
    udp_handlers_t::iterator it = m_oUDPHandlers.find(name);
    if (it != m_oUDPHandlers.end()) {
      got_callback = true;
      udp_recv_handler_t handler = it->second;
      handler(p_msg, pPoint, from);
    }

    pkg_handlers_t::iterator tit = m_oPkgHandlers.find(name);
    if (tit != m_oPkgHandlers.end()) {
      got_callback = true;
      pkg_recv_handler_t handler = tit->second;
      handler(p_msg);
    }
    if (!got_callback) {
      //                LOG_ERROR(frmwk)<<"protobuf_pkg::handle_udp_pkg(),
      //                a message with id: "<<name<<"has no handler!";
    }
  }
}
void send_message(tcp_connection_base *p_from,
                  const std::shared_ptr<Message> &p_msg) {
  package_ptr p(new protobuf_wrapper_pkg(p_msg));
  p_from->send(p);
}

void send_message(udp_point *p_from, const udp_endpoint &to,
                  const std::shared_ptr<Message> &p_msg) {
  package_ptr p(new protobuf_wrapper_pkg(p_msg));
  p_from->send(p, to);
}

} // namespace net
} // namespace ff
#endif
