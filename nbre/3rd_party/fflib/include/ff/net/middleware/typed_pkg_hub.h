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

#include "ff/net/common/archive.h"
#include "ff/net/common/common.h"
#include "ff/net/middleware/net_nervure.h"
#include "ff/net/middleware/package.h"
#include "ff/net/middleware/pkg_handler.h"
#include "ff/net/network/end_point.h"
#include <cassert>
#include <map>

namespace ff {
namespace net {
class typed_pkg_hub {
public:
  typed_pkg_hub();

  virtual ~typed_pkg_hub();

  tcp_pkg_handler *get_tcp_pkg_handler() { return m_pTCPPkgHandler; }

  udp_pkg_handler *get_udp_pkg_handler() { return m_pUDPPkgHandler; }

  template <typename PkgTy_>
  void to_recv_pkg(
      const typename pkg_recv_callback<package, PkgTy_>::pkg_recv_handler_t
          &handler) {
    std::shared_ptr<PkgTy_> pPkg(new PkgTy_());
    if (m_oPkgCreatorContainer.find(pPkg->type_id()) ==
        m_oPkgCreatorContainer.end())
      m_oPkgCreatorContainer.insert(std::make_pair(
          pPkg->type_id(), std::bind(package_new_wrapper<PkgTy_>::New)));
    pkg_recv_handler_t rh(
        std::bind(pkg_recv_callback<package, PkgTy_>::recv_handler,
                  std::placeholders::_1, handler));
    m_oPkgHandlers.insert(std::make_pair(pPkg->type_id(), rh));
  }

  template <class PkgTy_>
  void tcp_to_recv_pkg(
      const typename tcp_recv_callback<package, PkgTy_>::pkg_recv_handler_t
          &handler) {
    std::shared_ptr<PkgTy_> pPkg(new PkgTy_());
    if (m_oPkgCreatorContainer.find(pPkg->type_id()) ==
        m_oPkgCreatorContainer.end())
      m_oPkgCreatorContainer.insert(std::make_pair(
          pPkg->type_id(), std::bind(package_new_wrapper<PkgTy_>::New)));
    tcp_recv_handler_t rh(
        std::bind(tcp_recv_callback<package, PkgTy_>::recv_handler,
                  std::placeholders::_1, std::placeholders::_2, handler));
    m_oTCPHandlers.insert(std::make_pair(pPkg->type_id(), rh));
  }

  template <class PkgTy_>
  void udp_to_recv_pkg(
      typename udp_recv_callback<package, PkgTy_>::pkg_recv_handler_t handler) {
    std::shared_ptr<PkgTy_> pPkg(new PkgTy_());
    if (m_oPkgCreatorContainer.find(pPkg->type_id()) ==
        m_oPkgCreatorContainer.end())
      m_oPkgCreatorContainer.insert(std::make_pair(
          pPkg->type_id(), std::bind(package_new_wrapper<PkgTy_>::New)));
    udp_recv_handler_t rh(std::bind(
        udp_recv_callback<package, PkgTy_>::recv_handler, std::placeholders::_1,
        std::placeholders::_2, std::placeholders::_3, handler));
    m_oUDPHandlers.insert(std::make_pair(pPkg->type_id(), rh));
  }

protected:
  class typed_tcp_pkg_handler : public tcp_pkg_handler {
  public:
    typed_tcp_pkg_handler(typed_pkg_hub *pHub);

    virtual ~typed_tcp_pkg_handler();

    virtual void handle_pkg(tcp_connection_base *pFrom,
                            const shared_buffer &buf);

    virtual bool is_pkg_to_handle(uint32_t pkg_id);

  protected:
    typed_pkg_hub *m_pHub;
  };

  class typed_udp_pkg_handler : public udp_pkg_handler {
  public:
    typed_udp_pkg_handler(typed_pkg_hub *pHub);

    virtual ~typed_udp_pkg_handler();

    virtual void handle_pkg(udp_point *pPoint, const shared_buffer &buf,
                            const udp_endpoint &from);

    virtual bool is_pkg_to_handle(uint32_t pkg_id);

  protected:
    typed_pkg_hub *m_pHub;
  };

  void handle_tcp_pkg(tcp_connection_base *pFrom, const shared_buffer &buf);

  void handle_udp_pkg(udp_point *pPoint, const shared_buffer &buf,
                      const udp_endpoint &from);

  bool is_tcp_pkg_to_handle(uint32_t pkg_id);

  bool is_udp_pkg_to_handle(uint32_t pkg_id);

  package_ptr get_pkg(const shared_buffer &buf);

protected:
  typedef std::function<package_ptr()> pkg_creator_t;
  typedef std::function<void(package_ptr, tcp_connection_base *)>
      tcp_recv_handler_t;
  typedef std::function<void(package_ptr, udp_point *, udp_endpoint)>
      udp_recv_handler_t;
  typedef std::function<void(package_ptr)> pkg_recv_handler_t;
  typedef std::map<uint32_t, pkg_creator_t> pkg_creator_container_t;
  typedef std::map<uint32_t, tcp_recv_handler_t> tcp_handlers_t;
  typedef std::map<uint32_t, udp_recv_handler_t> udp_handlers_t;
  typedef std::map<uint32_t, pkg_recv_handler_t> pkg_handlers_t;

  tcp_pkg_handler *m_pTCPPkgHandler;
  udp_pkg_handler *m_pUDPPkgHandler;
  pkg_creator_container_t m_oPkgCreatorContainer;
  tcp_handlers_t m_oTCPHandlers;
  udp_handlers_t m_oUDPHandlers;
  pkg_handlers_t m_oPkgHandlers;
};

} // namespace net
} // end namespace ff
