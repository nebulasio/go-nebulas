//!this file is included by framework/event.h

//Users should definitely hook these critical events!!!
ENABLE_HOOK_EVENT(::ff::net::event::tcp_get_connection)
ENABLE_HOOK_EVENT(::ff::net::event::tcp_lost_connection)
ENABLE_HOOK_EVENT(::ff::net::event::pkg_send_failed)

//internal use
ENABLE_HOOK_EVENT(::ff::net::event::more::tcp_server_accept_connection)
ENABLE_HOOK_EVENT(::ff::net::event::more::tcp_server_accept_error)
ENABLE_HOOK_EVENT(::ff::net::event::more::tcp_client_get_connection_succ)
ENABLE_HOOK_EVENT(::ff::net::event::more::tcp_client_conn_error)
ENABLE_HOOK_EVENT(::ff::net::event::more::connect_sent_stream_error)
ENABLE_HOOK_EVENT(::ff::net::event::more::connect_recv_stream_error)

//!Please uncomment the following events to enable hooking them.

//ENABLE_HOOK_EVENT(::ffnet::event::more::tcp_server_start_listen)

//ENABLE_HOOK_EVENT(::ffnet::event::more::tcp_start_recv_stream)
//ENABLE_HOOK_EVENT(::ffnet::event::more::tcp_start_send_stream)
//ENABLE_HOOK_EVENT(::ffnet::event::more::tcp_client_start_connection)

//ENABLE_HOOK_EVENT(::ffnet::event::more::connect_sent_stream_succ)

//ENABLE_HOOK_EVENT(::ffnet::event::more::connect_recv_stream_succ)
