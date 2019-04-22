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
#include "cmd/dummy_neb/generator/checkers.h"
#include "cmd/dummy_neb/dummy_callback.h"
#include "core/net_ipc/nipc_pkg.h"

nbre_version_checker::nbre_version_checker() : checker_task_base() {}

nbre_version_checker::~nbre_version_checker() {}

void nbre_version_checker::check() {
  callback_handler::instance().add_version_handler(
      reinterpret_cast<uint64_t>(this),
      [this](uint64_t, uint32_t major, uint32_t minor, uint32_t patch) {
        nbre_version_ack ack;
        ack.set<p_holder>(0);
        ack.set<p_major>(major);
        ack.set<p_minor>(minor);
        ack.set<p_patch>(patch);
        auto result = ack.serialize_to_string();
        apply_result(result);
      });
  m_b_is_running = true;
  ipc_nbre_version(this, 0xfffful);
}
std::string nbre_version_checker::name() const { return "ipc_nbre_version"; }

nbre_nr_result_check::nbre_nr_result_check(const std::string &nr_handle)
    : checker_task_base(), m_nr_handle(nr_handle) {}

nbre_nr_result_check::~nbre_nr_result_check() {}

void nbre_nr_result_check::check() {
  callback_handler::instance().add_nr_result_handler(
      reinterpret_cast<uint64_t>(this),
      [this](uint64_t, const char *nr_result) {
        auto result = std::string(nr_result);
        apply_result(result);
      });
  m_b_is_running = true;
  ipc_nbre_nr_result_by_handle(this, m_nr_handle.c_str());
}

std::string nbre_nr_result_check::name() const {
  std::stringstream ss;
  ss << "ipc_nbre_nr_result(" << m_nr_handle << ")";
  return ss.str();
}

nbre_nr_handle_check::nbre_nr_handle_check(uint64_t start_block,
                                           uint64_t end_block)
    : checker_task_base(), m_start_block(start_block), m_end_block(end_block) {}

std::string nbre_nr_handle_check::name() const {
  std::stringstream ss;
  ss << "ipc_nbre_nr_handle(" << m_start_block << "," << m_end_block << ")";
  return ss.str();
}
nbre_nr_handle_check::~nbre_nr_handle_check() {}

void nbre_nr_handle_check::check() {
  callback_handler::instance().add_nr_handler(
      reinterpret_cast<uint64_t>(this),
      [this](uint64_t, const char *nr_handle) {
        if (!m_nr_result_checker) {
          m_nr_result_checker =
              std::make_shared<nbre_nr_result_check>(std::string(nr_handle));
          checker_tasks::instance().add_task(m_nr_result_checker);
        }
        apply_result(std::string(nr_handle));
      });
  m_b_is_running = true;
  ipc_nbre_nr_handle(this, m_start_block, m_end_block, 0);
}

nbre_dip_reward_check::nbre_dip_reward_check(uint64_t height)
    : checker_task_base(), m_height(height) {}

nbre_dip_reward_check::~nbre_dip_reward_check() {}

std::string nbre_dip_reward_check::name() const {
  std::stringstream ss;
  ss << "ipc_nbre_dip_reward(" << m_height << ")";
  return ss.str();
}

void nbre_dip_reward_check::check() {
  callback_handler::instance().add_dip_reward_handler(
      reinterpret_cast<uint64_t>(this),
      [this](uint64_t, const char *dip_reward) {
        apply_result(std::string(dip_reward));
      });
  m_b_is_running = true;
  ipc_nbre_dip_reward(this, m_height, 0);
}
