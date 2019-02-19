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
#include "ff/net/common/archive.h"
#include "ff/net/common/common.h"

namespace ff {
namespace net {
/*******************
 * You should declare your own package by deriving from this class.
 * Basically, you need to do two things,
 * 1. to specify the typeID, which should be unique;
 * 2. to implement the virtual function, archive.
 *    archive is used to serialize and deserialize the Package, and
 * a typical implementation is like this,
 *    virtual void archive(archive & ar)
 *    {
 *          ar.archive(m_strName);
 *    }
 *  You can find me examples about archive.
 */
class package {
protected:
  virtual void archive(marshaler &ar) = 0;

public:
  package(uint32_t typeID) : m_iTypeID(typeID) {}

  virtual ~package(){};

  uint32_t type_id() const { return m_iTypeID; }
  uint32_t &type_id() { return m_iTypeID; }

  void arch(marshaler &ar) {
    ar.archive(m_iTypeID);
    archive(ar);
  }

protected:
  uint32_t m_iTypeID;
}; // end class Package

typedef std::shared_ptr<package> package_ptr;
} // namespace net
} // namespace ff

