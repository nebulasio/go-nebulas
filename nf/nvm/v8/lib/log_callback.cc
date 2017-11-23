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
#include "logger.h"

#include <stdarg.h>

static LogFunc LOG = NULL;
static const char *LogLevelText[] = {"DEBUG", "WARN", "INFO", "ERROR"};

const char *GetLogLevelText(int level) {
  if (level > LogLevel::ERROR) {
    level = LogLevel::ERROR;
  } else if (level < LogLevel::DEBUG) {
    level = LogLevel::INFO;
  }

  return LogLevelText[level - 1];
};

void InitializeLogger(LogFunc log) { LOG = log; }

void NewNativeLogFunction(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
  globalTpl->Set(String::NewFromUtf8(isolate, "_native_log"),
                 FunctionTemplate::New(isolate, LogCallback),
                 static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                                PropertyAttribute::ReadOnly));
}

void LogCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  if (info.Length() < 2) {
    isolate->ThrowException(Exception::Error(
        String::NewFromUtf8(isolate, "_native_log: mssing params")));
    return;
  }

  Local<Value> level = info[0];
  if (!level->IsNumber()) {
    isolate->ThrowException(Exception::Error(
        String::NewFromUtf8(isolate, "_native_log: level must be number")));
    return;
  }

  Local<Value> msg = info[1];
  if (!msg->IsString()) {
    isolate->ThrowException(Exception::Error(
        String::NewFromUtf8(isolate, "_native_log: msg must be string")));
    return;
  }

  if (LOG == NULL) {
    return;
  }

  String::Utf8Value m(msg);
  LOG((level->ToInt32())->Int32Value(), *m);
}

void LogInfof(const char *format, ...) {
  if (LOG == NULL) {
    return;
  }

  va_list vl;
  va_start(vl, format);

  char *msg = NULL;
  vasprintf(&msg, format, vl);
  if (msg != NULL) {
    LOG(LogLevel::INFO, msg);
    free(msg);
  }

  va_end(vl);
}

void LogErrorf(const char *format, ...) {
  if (LOG == NULL) {
    return;
  }

  va_list vl;
  va_start(vl, format);

  char *msg = NULL;
  vasprintf(&msg, format, vl);
  if (msg != NULL) {
    LOG(LogLevel::ERROR, msg);
    free(msg);
  }

  va_end(vl);
}

void LogWarnf(const char *format, ...) {
  if (LOG == NULL) {
    return;
  }

  va_list vl;
  va_start(vl, format);

  char *msg = NULL;
  vasprintf(&msg, format, vl);
  if (msg != NULL) {
    LOG(LogLevel::WARN, msg);
    free(msg);
  }

  va_end(vl);
}

void LogDebugf(const char *format, ...) {
  if (LOG == NULL) {
    return;
  }

  va_list vl;
  va_start(vl, format);

  char *msg = NULL;
  vasprintf(&msg, format, vl);
  if (msg != NULL) {
    LOG(LogLevel::DEBUG, msg);
    free(msg);
  }

  va_end(vl);
}

void LogFatalf(const char *format, ...) {
  if (LOG == NULL) {
    return;
  }

  va_list vl;
  va_start(vl, format);

  char *msg = NULL;
  vasprintf(&msg, format, vl);
  if (msg != NULL) {
    LOG(LogLevel::ERROR, msg);
    free(msg);
  }

  va_end(vl);
}
