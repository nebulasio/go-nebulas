#include "random.h"
#include "global.h"
#include "../engine.h"
#include "instruction_counter.h"
#include "logger.h"

static GetTxRandomFunc sRandomDelegate = NULL;

// void NewNativeRandomFunction(Isolate *isolate,
//                               Local<ObjectTemplate> globalTpl) {
//   globalTpl->Set(String::NewFromUtf8(isolate, "_native_random"),
//                  FunctionTemplate::New(isolate, RandomCallback),
//                  static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
//                                                 PropertyAttribute::ReadOnly));
// }

void NewRandomInstance(Isolate *isolate, Local<Context> context,
                           void *handler) {
  Local<ObjectTemplate> blockTpl = ObjectTemplate::New(isolate);
//   Local<Object> blockTpl = context->Global();
  blockTpl->SetInternalFieldCount(1);

  blockTpl->Set(String::NewFromUtf8(isolate, "random"),
                FunctionTemplate::New(isolate, RandomCallback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly));

  Local<Object> instance = blockTpl->NewInstance(context).ToLocalChecked();
  instance->SetInternalField(0, External::New(isolate, handler));
  
  context->Global()->DefineOwnProperty(
      context, String::NewFromUtf8(isolate, "_native_math"), instance,
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly | PropertyAttribute::DontEnum));
}

void RandomCallback(const v8::FunctionCallbackInfo<v8::Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));
  if (info.Length() != 0) {
    isolate->ThrowException(
        Exception::Error(String::NewFromUtf8(isolate, "require random args err")));
    return;
  }
//   size_t cnt = 0;

  // char *value = sRandomDelegate(handler->Value());
  // if (value == NULL) {
  //   info.GetReturnValue().SetNull();
  // } else {
  //   info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
  //   free(value);
  // }
  size_t cnt = 0;
  char *result = NULL;
  char *exceptionInfo = NULL;

  int err = sRandomDelegate(handler->Value(), &cnt, &result, &exceptionInfo);
  DEAL_ERROR_FROM_GOLANG(err);

  if (result != NULL) {
    free(result);
  }

  if (exceptionInfo != NULL) {
    free(exceptionInfo);
  }
  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

void InitializeRandom(GetTxRandomFunc delegate) {
  sRandomDelegate = delegate;
}
