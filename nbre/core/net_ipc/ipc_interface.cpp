// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//
#include "core/net_ipc/ipc_interface.h"
#include "common/configuration.h"
#include "core/net_ipc/nipc_pkg.h"
#include "core/net_ipc/server/nipc_server.h"
#include <ff/network.h>

std::shared_ptr<neb::core::nipc_server> _ipc;

int start_nbre_ipc(nbre_params_t params) {
  try {
    LOG(INFO) << "log-to-stderr: " << neb::glog_log_to_stderr;
    if (neb::glog_log_to_stderr) {
      FLAGS_log_dir = params.m_nbre_log_dir;
      LOG(INFO) << "log dir server " << FLAGS_log_dir;
      google::InitGoogleLogging("nbre-server");
    }

    _ipc = std::make_shared<neb::core::nipc_server>();
    LOG(INFO) << "ipc server construct";
    _ipc->init_params(params);

    LOG(INFO) << "ipc listen " << params.m_nipc_listen;
    LOG(INFO) << "ipc port " << params.m_nipc_port;
    if (_ipc->start()) {
      LOG(INFO) << "start nbre succ";
      return ipc_status_succ;
    } else {
      LOG(ERROR) << "start nbre failed";
      return ipc_status_fail;
    }
  } catch (const std::exception &e) {
    LOG(ERROR) << "start nbre got exception " << typeid(e).name() << ":"
               << e.what();
    return ipc_status_exception;
  } catch (...) {
    LOG(ERROR) << "start nbre got unknown exception ";
    return ipc_status_exception;
  }
}

void nbre_ipc_shutdown() {
  if (!_ipc)
    return;
  _ipc->shutdown();
  _ipc.reset();
}

//////////////////////////////////////
template <typename P1, typename... Params> struct set_bind_helper {
  template <typename PkgPtrType, typename A1, typename... Args>
  static void set(PkgPtrType &pkg, A1 a, Args... args) {
    pkg->template set<P1>(a);
    set_bind_helper<Params...>::set(pkg, args...);
  }
  template <typename PkgPtrType, typename... Args>
  static void set(PkgPtrType &pkg, void *a, Args... args) {
    pkg->template set<P1>(reinterpret_cast<uint64_t>(a));
    set_bind_helper<Params...>::set(pkg, args...);
  }
};

template <typename P1> struct set_bind_helper<P1> {
  template <typename PkgPtrType, typename A1>
  static void set(PkgPtrType &pkg, A1 a) {
    pkg->template set<P1>(a);
  }
  template <typename PkgPtrType> static void set(PkgPtrType &pkg, void *a) {
    pkg->template set<P1>(reinterpret_cast<uint64_t>(a));
  }
};

template <typename PkgType, typename... Params> struct ipc_call {
  template <typename... Args> static int bind(Args... args) {
    std::shared_ptr<PkgType> pkg(new PkgType());
    set_bind_helper<Params...>::set(pkg, args...);
    void *holder = reinterpret_cast<void *>(pkg->template get<p_holder>());
    if (!_ipc)
      return ipc_status_fail;
    return _ipc->send_api_pkg<PkgType>(holder, pkg);
  }
};
///////////////////////////////////////

template <typename IT,
          typename DT = typename ::ff::util::internal::nt_traits<IT>::type>
struct get_param_helper {
  template <typename PkgPtrType> static DT get(const PkgPtrType &pkg) {
    return pkg->template get<IT>();
  };
};

template <>
struct get_param_helper<
    p_holder, typename ::ff::util::internal::nt_traits<p_holder>::type> {
  template <typename PkgPtrType> static void *get(const PkgPtrType &pkg) {
    return reinterpret_cast<void *>(pkg->template get<p_holder>());
  };
};
template <>
struct get_param_helper<
    p_nr_result, typename ::ff::util::internal::nt_traits<p_nr_result>::type> {
  template <typename PkgPtrType> static std::string get(const PkgPtrType &pkg) {
    std::string r = pkg->template get<p_nr_result>();

    nr_result nr;
    nr.deserialize_from_string(r);
    std::string t = ::neb::core::convert_nr_result_to_json(nr);
    return t;
  }
};
// template <typename IT> struct get_param_helper<IT, std::string> {
// template <typename PkgPtrType> static const char *get(const PkgPtrType &pkg)
// { auto t = pkg->template get<IT>().c_str(); LOG(INFO) << "get param: " << t
//<< ", should be: " << pkg->template get<IT>();
// return t;
//};
//};

template <typename PkgType, typename... Params>
struct ipc_callback {

  // template <typename T, typename PkgPtrType>
  // static auto get_param_for_callback(const PkgPtrType &pkg) {
  // return get_param_helper<T>::get(pkg);
  //}

  template <typename T> static auto get_param_for_callback(const T &val) {
    return val;
  }
  static const char *get_param_for_callback(const std::string &val) {
    return val.c_str();
  }

  template <typename T1, typename Func>
  static void callback_invoke(Func &&f, PkgType *pkg,
                              enum ipc_status_code code) {
    auto t1 = get_param_helper<T1>::get(pkg);
    f(code, get_param_for_callback(t1));
  }
  template <typename T1, typename T2, typename Func>
  static void callback_invoke(Func &&f, PkgType *pkg,
                              enum ipc_status_code code) {
    auto t1 = get_param_helper<T1>::get(pkg);
    auto t2 = get_param_helper<T2>::get(pkg);
    f(code, get_param_for_callback(t1), get_param_for_callback(t2));
  }
  template <typename T1, typename T2, typename T3, typename Func>
  static void callback_invoke(Func &&f, PkgType *pkg,
                              enum ipc_status_code code) {
    auto t1 = get_param_helper<T1>::get(pkg);
    auto t2 = get_param_helper<T2>::get(pkg);
    auto t3 = get_param_helper<T3>::get(pkg);
    f(code, get_param_for_callback(t1), get_param_for_callback(t2),
      get_param_for_callback(t3));
  }
  template <typename T1, typename T2, typename T3, typename T4, typename Func>
  static void callback_invoke(Func &&f, PkgType *pkg,
                              enum ipc_status_code code) {
    auto t1 = get_param_helper<T1>::get(pkg);
    auto t2 = get_param_helper<T2>::get(pkg);
    auto t3 = get_param_helper<T3>::get(pkg);
    auto t4 = get_param_helper<T4>::get(pkg);
    f(code, get_param_for_callback(t1), get_param_for_callback(t2),
      get_param_for_callback(t3), get_param_for_callback(t4));
  }
  template <typename T1, typename T2, typename T3, typename T4, typename T5,
            typename Func>
  static void callback_invoke(Func &&f, PkgType *pkg,
                              enum ipc_status_code code) {
    auto t1 = get_param_helper<T1>::get(pkg);
    auto t2 = get_param_helper<T2>::get(pkg);
    auto t3 = get_param_helper<T3>::get(pkg);
    auto t4 = get_param_helper<T4>::get(pkg);
    auto t5 = get_param_helper<T5>::get(pkg);
    f(code, get_param_for_callback(t1), get_param_for_callback(t2),
      get_param_for_callback(t3), get_param_for_callback(t4),
      get_param_for_callback(t5));
  }
  template <typename T1, typename T2, typename T3, typename T4, typename T5,
            typename T6, typename Func>
  static void callback_invoke(Func &&f, PkgType *pkg,
                              enum ipc_status_code code) {
    auto t1 = get_param_helper<T1>::get(pkg);
    auto t2 = get_param_helper<T2>::get(pkg);
    auto t3 = get_param_helper<T3>::get(pkg);
    auto t4 = get_param_helper<T4>::get(pkg);
    auto t5 = get_param_helper<T5>::get(pkg);
    auto t6 = get_param_helper<T6>::get(pkg);
    f(code, get_param_for_callback(t1), get_param_for_callback(t2),
      get_param_for_callback(t3), get_param_for_callback(t4),
      get_param_for_callback(t5), get_param_for_callback(t6));
  }
  template <typename Func, typename... Args>
  static void bind(Func &&func, Args... args) {
    neb::core::ipc_callback_holder::instance().add_callback(
        PkgType().type_id(),
        [func, args...](enum ipc_status_code code, ff::net::package *pkg) {
          PkgType *ack = (PkgType *)pkg;
          if (code == ipc_status_succ) {
            callback_invoke<Params...>(func, ack, code);
          } else {
            neb::core::issue_callback_with_error(func, code);
          }
        });
  }
};

////////////////////////////////////////////////
int ipc_nbre_version(void *holder, uint64_t height) {
  return ipc_call<nbre_version_req, p_holder, p_height>::bind(holder, height);
}
void set_recv_nbre_version_callback(nbre_version_callback_t func) {
  ipc_callback<nbre_version_ack, p_holder, p_major, p_minor, p_patch>::bind(
      func, _2, _3, _4, _5);
}
int ipc_nbre_ir_list(void *holder) {
  return ipc_call<nbre_ir_list_req, p_holder>::bind(holder);
}

void set_recv_nbre_ir_list_callback(nbre_ir_list_callback_t func) {
  ipc_callback<nbre_ir_list_ack, p_holder, p_ir_name_list>::bind(func, _2, _3);
}

// interface ipc_nbre_ir_versions
int ipc_nbre_ir_versions(void *holder, const char *ir_name) {
  return ipc_call<nbre_ir_versions_req, p_holder, p_ir_name>::bind(holder,
                                                                   ir_name);
}
void set_recv_nbre_ir_versions_callback(nbre_ir_versions_callback_t func) {
  ipc_callback<nbre_ir_versions_ack, p_holder, p_ir_versions>::bind(func, _2,
                                                                    _3);
}

// interface get nr handle
int ipc_nbre_nr_handle(void *holder, uint64_t start_block, uint64_t end_block,
                       uint64_t nr_version) {
  return ipc_call<nbre_nr_handle_req, p_holder, p_start_block, p_end_block,
                  p_nr_version>::bind(holder, start_block, end_block,
                                      nr_version);
}
void set_recv_nbre_nr_handle_callback(nbre_nr_handle_callback_t func) {
  ipc_callback<nbre_nr_handle_ack, p_holder, p_nr_handle>::bind(func, _2, _3);
}

// interface get nr result by handle
int ipc_nbre_nr_result_by_handle(void *holder, const char *nr_handle) {
  return ipc_call<nbre_nr_result_by_handle_req, p_holder, p_nr_handle>::bind(
      holder, nr_handle);
}
void set_recv_nbre_nr_result_by_handle_callback(
    nbre_nr_result_by_handle_callback_t func) {
  ipc_callback<nbre_nr_result_by_handle_ack, p_holder, p_nr_result>::bind(
      func, _2, _3);
}

// interface get nr result by height
int ipc_nbre_nr_result_by_height(void *holder, uint64_t height) {
  return ipc_call<nbre_nr_result_by_height_req, p_holder, p_height>::bind(
      holder, height);
}
void set_recv_nbre_nr_result_by_height_callback(
    nbre_nr_result_by_height_callback_t func) {
  ipc_callback<nbre_nr_result_by_height_ack, p_holder, p_nr_result>::bind(
      func, _2, _3);
}

// interface get nr sum
int ipc_nbre_nr_sum(void *holder, uint64_t height) {
  return ipc_call<nbre_nr_sum_req, p_holder, p_height>::bind(holder, height);
}
void set_recv_nbre_nr_sum_callback(nbre_nr_sum_callback_t func) {
  ipc_callback<nbre_nr_sum_ack, p_holder, p_nr_sum>::bind(func, _2, _3);
}

// interface get dip reward
int ipc_nbre_dip_reward(void *holder, uint64_t height, uint64_t version) {
  return ipc_call<nbre_dip_reward_req, p_holder, p_height, p_version>::bind(
      holder, height, version);
}
void set_recv_nbre_dip_reward_callback(nbre_dip_reward_callback_t func) {
  ipc_callback<nbre_dip_reward_ack, p_holder, p_dip_reward>::bind(func, _2, _3);
}

// interface send ir transactions
std::unique_ptr<std::vector<std::string>> _txs_ptr;
uint64_t _height;
int ipc_nbre_ir_transactions_create(void *holder, uint64_t height) {
  _txs_ptr = std::make_unique<std::vector<std::string>>();
  _height = height;
  return ipc_status_succ;
}
int ipc_nbre_ir_transactions_append(void *holder, uint64_t height,
                                    const char *tx_bytes,
                                    int32_t tx_bytes_len) {
  if (height != _height) {
    return ipc_status_fail;
  }
  _txs_ptr->push_back(std::string(tx_bytes, tx_bytes_len));
  return ipc_status_succ;
}
int ipc_nbre_ir_transactions_send(void *holder, uint64_t height) {
  if (height != _height) {
    return ipc_status_fail;
  }
  return ipc_call<nbre_ir_transactions_req, p_holder, p_height,
                  p_ir_transactions>::bind(holder, _height, *_txs_ptr);
}

