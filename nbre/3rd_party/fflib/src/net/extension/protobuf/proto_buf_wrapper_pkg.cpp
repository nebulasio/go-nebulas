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
#include "ff/net/extension/protobuf/proto_buf_wrapper_pkg.h"
#include "ff/net/common/common.h"
#ifdef PROTO_BUF_SUPPORT

namespace ff {
namespace net {
protobuf_wrapper_pkg::protobuf_wrapper_pkg()
    : package(protobuf_wrapper_pkg_type) {}
protobuf_wrapper_pkg::protobuf_wrapper_pkg(const std::string &strPBMessageName)
    : package(protobuf_wrapper_pkg_type),
      m_strProtoBufMsgName(strPBMessageName) {}

protobuf_wrapper_pkg::protobuf_wrapper_pkg(std::shared_ptr<Message> pMsg)
    : package(protobuf_wrapper_pkg_type),
      m_strProtoBufMsgName(pMsg->GetDescriptor()->full_name()), m_pPBMsg(pMsg) {

}

std::shared_ptr<Message> protobuf_wrapper_pkg::protobuf_message() const {
  return m_pPBMsg;
}

void protobuf_wrapper_pkg::archive(marshaler &ar) {
  // optimizing for each archiver
  ar.archive(m_strProtoBufMsgName);

  switch (ar.get_marshaler_type()) {
  case marshaler::deseralizer: {
    std::string buf;
    ar.archive(buf);
    m_pPBMsg = create_message(m_strProtoBufMsgName);
    assert(m_pPBMsg != NULL && "Can't find message in protobuf");
    m_pPBMsg->ParseFromString(buf);
    break;
  }
  case marshaler::seralizer: {
    std::string buf;
    m_pPBMsg->SerializeToString(&buf);
    ar.archive(buf);
    break;
  }
  case marshaler::length_retriver: {
    assert(m_pPBMsg != NULL && "Didn't set m_pPBMsg yet!");
    std::string buf(m_pPBMsg->ByteSize(), '0');
    ar.archive(buf);
    break;
  }
  }
}

protobuf_wrapper_pkg::message_ptr
protobuf_wrapper_pkg::create_message(const std::string &typeName) {
  message_ptr message;
  const Descriptor *descriptor =
      DescriptorPool::generated_pool()->FindMessageTypeByName(typeName);
  if (descriptor) {
    const Message *prototype =
        MessageFactory::generated_factory()->GetPrototype(descriptor);
    if (prototype) {
      message = message_ptr(prototype->New());
    }
  }
  return message;
}

} // namespace net
} // namespace ff
#endif
