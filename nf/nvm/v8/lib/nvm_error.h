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

#ifndef _NEBULAS_NF_NVM_V8_ERROR_H_
#define _NEBULAS_NF_NVM_V8_ERROR_H_

/*
success or crash
*/
#define REPORT_UNEXPECTED_ERR() do{ \
  if (NULL == isolate) {\
    LogFatalf("Unexpected Error: invalid argument, ioslate is NULL");\
  }\
  Local<Context> context = isolate->GetCurrentContext();\
  V8Engine *e = GetV8EngineInstance(context);\
  if (NULL == e) {\
    LogFatalf("Unexpected Error: failed to get V8Engine");\
  }\
  TerminateExecution(e);\
  e->is_unexpected_error_happen = true;\
} while(0)

#define DEAL_ERROR_FROM_GOLANG(err) do {\
  if (NVM_UNEXPECTED_ERR == err || (NVM_EXCEPTION_ERR == err && NULL == exceptionInfo) ||\
    (NVM_SUCCESS == err && NULL == result)) {\
    info.GetReturnValue().SetNull();\
    REPORT_UNEXPECTED_ERR();\
  } else if (NVM_EXCEPTION_ERR == err) {\
    isolate->ThrowException(String::NewFromUtf8(isolate, exceptionInfo));\
  } else if (NVM_SUCCESS == err) {\
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, result));\
  } else {\
    info.GetReturnValue().SetNull();\
    REPORT_UNEXPECTED_ERR();\
  }\
} while(0)

enum nvmErrno {
  NVM_SUCCESS = 0,
  NVM_EXCEPTION_ERR = -1,
  NVM_MEM_LIMIT_ERR = -2,
  NVM_GAS_LIMIT_ERR = -3,
  NVM_UNEXPECTED_ERR = -4,
  NVM_EXE_TIMEOUT_ERR = -5,
  NVM_INNER_EXE_ERR = -6,
};

#endif //_NEBULAS_NF_NVM_V8_ENGINE_ERROR_H_