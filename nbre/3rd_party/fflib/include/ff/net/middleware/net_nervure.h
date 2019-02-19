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
#include "ff/net/middleware/event_handler.h"
#include "ff/net/middleware/pkg_handler.h"
#include "ff/net/middleware/pkg_packer.h"
#include "ff/net/network/end_point.h"
#include "ff/net/network/events.h"
#include "ff/net/network/tcp_connection_base.h"
#include <condition_variable>
#include <mutex>

namespace ff {
namespace net {
class tcp_server;

class udp_point;

typedef std::shared_ptr<tcp_server> tcp_server_ptr;
typedef std::shared_ptr<udp_point> udp_point_ptr;

using boost::asio::io_service;
using namespace event;

class net_nervure {
public:
  net_nervure(net_mode nm = real_net);

  void add_pkg_handler(tcp_pkg_handler *p_tcp_handler,
                       udp_pkg_handler *p_udp_handler);

  template <class HT_> void add_pkg_hub(HT_ &ht) {
    if (ht.get_tcp_pkg_handler() != NULL) {
      m_pTCPHandler.push_back(ht.get_tcp_pkg_handler());
    }
    if (ht.get_udp_pkg_handler() != NULL) {
      m_pUDPHandler.push_back(ht.get_udp_pkg_handler());
    }
  }

  virtual ~net_nervure();

  void run();

  //! This is thread safe!
  void stop();

  tcp_server *add_tcp_server(const std::string &ip, uint16_t iTCPPort);

  tcp_server *add_tcp_server(const tcp_endpoint &ep);

  udp_point *add_udp_point(const std::string &ip, uint16_t iUDPPort);

  udp_point *add_udp_point(const udp_endpoint &ep);

  tcp_connection_base *add_tcp_client(const std::string &ip, uint16_t iTCPPort);

  tcp_connection_base *add_tcp_client(const tcp_endpoint &ep);

  inline io_service &ioservice() { return m_oIOService; }

  inline pkg_packer *get_pkg_packer() { return m_pBS; }

  inline event_handler *get_event_handler() { return m_pEH; }

  inline std::vector<tcp_pkg_handler *> get_tcp_pkg_handler() {
    return m_pTCPHandler;
  }

  inline std::vector<udp_pkg_handler *> get_udp_pkg_handler() {
    return m_pUDPHandler;
  }

protected:
  void on_tcp_server_accept_connect(tcp_connection_base_ptr pConn);

  void on_tcp_client_get_connect(tcp_connection_base *pClient);

  void on_conn_recv_or_send_error(tcp_connection_base *pConn,
                                  boost::system::error_code ec);

  void internal_stop();

protected:
  pkg_packer *m_pBS;
  event_handler *m_pEH;
  std::vector<tcp_pkg_handler *> m_pTCPHandler;
  std::vector<udp_pkg_handler *> m_pUDPHandler;
  io_service m_oIOService;

  net_mode mi_mode;
  typedef std::vector<tcp_server_ptr> tcp_servers_t;
  typedef std::vector<tcp_connection_base_ptr> tcp_clients_t;
  typedef std::vector<udp_point_ptr> udp_points_t;
  typedef std::vector<tcp_connection_base_ptr> tcp_connections_t;

  tcp_servers_t m_oServers;
  tcp_clients_t m_oClients;
  udp_points_t m_oUDPPoints;
  tcp_connections_t m_oTCPConnections;
  bool m_safe_to_stop;
  std::mutex m_stop_mutex;
  std::condition_variable m_stop_cond;
  std::thread::id m_io_service_thrd;
}; // end class

} // namespace net
} // namespace ff
