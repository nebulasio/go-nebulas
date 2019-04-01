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

//! This file is to implement archive for different types.

template <class T>
typename std::enable_if<std::is_arithmetic<T>::value, void>::type
archive(T &data) {
  switch (m_iAT) {
  case seralizer:
    std::memcpy(m_pWriteBuf + m_iBase, (const char *)&data, sizeof(T));
    m_iBase += sizeof(T);
    break;
  case deseralizer:
    std::memcpy((char *)&data, m_pReadBuf + m_iBase, sizeof(T));
    m_iBase += sizeof(T);
    break;
  case length_retriver:
    m_iBase += sizeof(T);
    break;
  }
}

void archive(std::string &s);

template <class T, size_t N>
typename std::enable_if<std::is_arithmetic<T>::value, void>::type
archive(T (&data)[N]) {
  switch (m_iAT) {
  case seralizer:
    std::memcpy(m_pWriteBuf + m_iBase, (const char *)&data, sizeof(T) * N);
    m_iBase += sizeof(T) * N;
    break;
  case deseralizer:
    std::memcpy((char *)&data, m_pReadBuf + m_iBase, sizeof(T) * N);
    m_iBase += sizeof(T) * N;
    break;
  case length_retriver:
    m_iBase += sizeof(T) * N;
    break;
  }
}

template <class T, size_t N>
typename std::enable_if<!std::is_arithmetic<T>::value, void>::type
archive(T (&data)[N]) {
  for (size_t i = 0; i < N; i++) {
    archive(data[i]);
  }
}

template <class T>
typename std::enable_if<std::is_arithmetic<T>::value, void>::type
archive(T *&data, size_t count) {
  switch (m_iAT) {
  case seralizer:
    std::memcpy(m_pWriteBuf + m_iBase, (const char *)data, sizeof(T) * count);
    m_iBase += sizeof(T) * count;
    break;
  case deseralizer:
    std::memcpy((char *)data, m_pReadBuf + m_iBase, sizeof(T) * count);
    m_iBase += sizeof(T) * count;
    break;
  case length_retriver:
    m_iBase += sizeof(T) * count;
    break;
  }
}

template <class T>
typename std::enable_if<std::is_arithmetic<T>::value, void>::type
archive(std::vector<T> &data) {
  size_t count = data.size();
  archive(count);
  switch (m_iAT) {
  case seralizer:
    std::memcpy(m_pWriteBuf + m_iBase, (const char *)&data.front(),
                sizeof(T) * count);
    m_iBase += count * sizeof(T);
    break;
  case deseralizer:
    data.clear();
    for (int i = 0; i < count; ++i) {
      T d;
      archive(d);
      data.push_back(d);
    }
    break;
  case length_retriver:
    m_iBase += sizeof(T) * count;
    break;
  }
}

template <class T>
typename std::enable_if<!std::is_arithmetic<T>::value, void>::type
archive(std::vector<T> &data) {
  size_t count = data.size();
  archive(count);
  switch (m_iAT) {
  case seralizer:
    for (size_t i = 0; i < data.size(); ++i) {
      archive(data[i]);
    }
    break;
  case deseralizer:
    data.clear();
    for (int i = 0; i < count; ++i) {
      T d;
      archive(d);
      data.push_back(d);
    }
    break;
  case length_retriver:
    for (size_t i = 0; i < data.size(); ++i) {
      archive(data[i]);
    }
    break;
  }
}

