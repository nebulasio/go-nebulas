// Copyright (C) 2017 go-nebulas authors
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
//

#include "log_callback.h"

static LogFunc logFunc = NULL;
static const char *LogLevelText[] = {"DEBUG", "WARN", "INFO", "ERROR"};

const char *GetLogLevelText(int level) {
  if (level > LogLevel::ERROR) {
    level = LogLevel::ERROR;
  } else if (level < LogLevel::DEBUG) {
    level = LogLevel::INFO;
  }

  return LogLevelText[level - 1];
};

void SetLogFunc(LogFunc f) { logFunc = f; }

void logCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  if (info.Length() < 2) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "native_log mssing params"));
    return;
  }

  Local<Value> level = info[0];
  if (!level->IsNumber()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "level must be number"));
    return;
  }

  Local<Value> msg = info[1];
  if (!msg->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "msg must be string"));
    return;
  }

  if (logFunc == NULL) {
    return;
  }

  String::Utf8Value m(msg);
  logFunc((level->ToInt32())->Int32Value(), *m);
}
