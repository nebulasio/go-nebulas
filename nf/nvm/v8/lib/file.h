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

#ifndef _NEBULAS_NF_NVM_V8_LIB_FILE_H_
#define _NEBULAS_NF_NVM_V8_LIB_FILE_H_

#include <stddef.h>

#define MAX_PATH_LEN    1024
#define MAX_VERSION_LEN 64
#define MAX_VERSIONED_PATH_LEN 1088

#ifdef _WIN32
    #define FILE_SEPARATOR "\\"
    #define LIB_DIR     "\\lib"
    #define EXECUTION_FILE  "\\execution_env.js"
#else
    #define FILE_SEPARATOR "/"
    #define LIB_DIR     "/lib"
    #define EXECUTION_FILE  "/execution_env.js"
#endif //WIN32

char *readFile(const char *filepath, size_t *size);

bool isFile(const char *file);
bool getCurAbsolute(char *curCwd, int len);

#endif // _NEBULAS_NF_NVM_V8_LIB_FILE_H_
