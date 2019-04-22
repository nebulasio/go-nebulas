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
#include "ff/net/common/common.h"
#include "ff/net/common/shared_buffer.h"
#include "ff/net/middleware/package.h"
#include <list>

namespace ff {
namespace net {
class net_buffer;
class pkg_packer {
public:
  virtual ~pkg_packer(){};
  virtual std::list<shared_buffer> split(net_buffer &oRecvBuffer) = 0;
  virtual void pack(net_buffer &oSendBuffer, const char *pBuf, size_t len) = 0;
  virtual void pack(net_buffer &oSendBuffer, const package_ptr &pkg) = 0;
}; // end class PkgPacker
typedef std::shared_ptr<pkg_packer> pkg_packer_ptr;

class length_packer : public pkg_packer {
public:
  virtual ~length_packer();
  virtual std::list<shared_buffer> split(net_buffer &oRecvBuffer);
  virtual void pack(net_buffer &oSendBuffer, const char *pBuf, size_t len);
  virtual void pack(net_buffer &oSendBuffer, const package_ptr &pkg);
};

} // namespace net
} // namespace ff
