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
#ifdef PROTO_BUF_SUPPORT

#include "ff/net/common/archive.h"
#include "ff/net/common/common.h"
#include "ff/net/middleware/package.h"

#include <google/protobuf/descriptor.h>
#include <google/protobuf/message.h>

namespace ff {
namespace net {
using google::protobuf::DescriptorPool;
using google::protobuf::MessageFactory;
using google::protobuf::Message;
using google::protobuf::Descriptor;

class protobuf_wrapper_pkg : public package {
public:
  protobuf_wrapper_pkg();

  protobuf_wrapper_pkg(const std::string &strPBMessageName);

  protobuf_wrapper_pkg(std::shared_ptr<google::protobuf::Message> pMsg);

  std::shared_ptr<Message> protobuf_message() const;

  virtual void archive(marshaler &ar);

protected:
  typedef std::shared_ptr<Message> message_ptr;
  message_ptr create_message(const std::string &typeName);

protected:
  std::string m_strProtoBufMsgName;
  message_ptr m_pPBMsg;
  // char *        m_pProtoBufMsg;
}; // end class ProtoBufWrapperPkg

} // namespace net
} // namespace ff
#endif
