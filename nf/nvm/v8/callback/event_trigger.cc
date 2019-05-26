// Copyright (C) 2017-2019 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify it under the terms of the GNU General Public License as published by
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


#include "event_trigger.h"

void EventTrigger(void *handler, const char *topic, const char *data, size_t *cnt){

  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(EVENT_TRIGGER_FUNC));
  res->add_func_params(std::string(topic));
  res->add_func_params(std::string(data));

  LogInfof("[Event] [%s] %s\n", topic, data);

  const NVMCallbackResult *callback_res = SNVM::DataExchangeCallback(handler, res);
  *cnt = (size_t)std::stoll(callback_res->result());
  if(callback_res != nullptr)
    delete callback_res;
}