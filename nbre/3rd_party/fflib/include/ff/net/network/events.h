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
#include "ff/net/network/tcp_client.h"
#include "ff/net/network/tcp_connection_base.h"
#include "ff/net/network/tcp_server.h"
#include "ff/net/network/udp_point.h"

namespace ff {
namespace net {
namespace event {
struct tcp_get_connection {
  const static size_t identifier = 1;
  typedef std::function<void(tcp_connection_base *)> Handler_t;
};
// end tcp_get_connection
struct tcp_lost_connection {
  const static size_t identifier = 2;
  typedef std::function<void(tcp_connection_base *)> Handler_t;
}; // end tcp_get_connection

//! This happens because the socked is invalid due to lost connection or errors
struct tcp_pkg_send_failed {
  const static size_t identifier = 3;
  typedef ff::net::package_ptr package_ptr;
  typedef std::function<void(tcp_connection_base *, package_ptr)> Handler_t;
};

struct udp_send_recv_exception {
  const static size_t identifier = 4;
  // TODO, what's the event_handler we need?
}; // end udp_send_recv_exception

namespace more {
struct tcp_server_start_listen {
  const static size_t identifier = 5;
  typedef std::function<void(tcp_endpoint)> Handler_t;
}; // end struct tcp_server_start_listen

struct tcp_server_accept_connection {
  const static size_t identifier = 6;
  typedef std::function<void(ff::net::tcp_connection_base_ptr)> Handler_t;
};
// end struct tcp_server_accept_connection
struct tcp_server_accept_error {
  const static size_t identifier = 7;
  typedef boost::system::error_code error_code;
  typedef std::function<void(tcp_endpoint, error_code)> Handler_t;
};
// end tcp_server_accept_error
struct tcp_start_recv_stream {
  const static size_t identifier = 8;
  typedef std::function<void(tcp_endpoint, tcp_endpoint)> Handler_t;
};
// end tcp_start_recv_stream
struct tcp_start_send_stream {
  const static size_t identifier = 9;
  typedef std::function<void(ff::net::tcp_connection_base *, const char *,
                             size_t)>
      Handler_t;
};
// end tcp_start_send_stream
struct tcp_client_start_connection {
  const static size_t identifier = 10;
  typedef std::function<void(tcp_endpoint)> Handler_t;
};
// end tcp_client_start_connection
struct tcp_client_get_connection_succ {
  const static size_t identifier = 11;
  typedef std::function<void(ff::net::tcp_connection_base *)> Handler_t;
};
// end tcp_client_get_connection
struct tcp_client_conn_error {
  const static size_t identifier = 12;
  typedef boost::system::error_code error_code;
  typedef std::function<void(ff::net::tcp_connection_base *, error_code)>
      Handler_t;
};
// end tcp_client_conn_error
struct tcp_send_stream_succ {
  const static size_t identifier = 13;
  typedef std::function<void(ff::net::tcp_connection_base *, size_t)> Handler_t;
};
// end
struct tcp_send_stream_error {
  const static size_t identifier = 14;
  typedef boost::system::error_code error_code;
  typedef std::function<void(ff::net::tcp_connection_base *, error_code)>
      Handler_t;
};
// end connect_sent_stream_error
struct tcp_recv_stream_succ {
  const static size_t identifier = 15;
  typedef std::function<void(ff::net::tcp_connection_base *, size_t)> Handler_t;
};
// end connect_recv_stream_succ
struct tcp_recv_stream_error {
  const static size_t identifier = 16;
  typedef boost::system::error_code error_code;
  typedef std::function<void(ff::net::tcp_connection_base *, error_code)>
      Handler_t;
}; // end connect_recv_stream_error

struct udp_send_data_succ {
  const static size_t identifier = 17;
  typedef std::function<void(udp_point *, size_t)> Handler_t;
};
struct udp_send_data_error {
  const static size_t identifier = 18;
  typedef boost::system::error_code error_code;
  typedef std::function<void(udp_point *, error_code)> Handler_t;
};

struct udp_recv_data_succ {
  const static size_t identifier = 19;
  typedef std::function<void(udp_point *, size_t)> Handler_t;
};

struct udp_recv_data_error {
  const static size_t identifier = 20;
  typedef std::function<void(udp_point *, boost::system::error_code)> Handler_t;
};

} // end namespace more
} // namespace event

} // namespace net
} // namespace ff

