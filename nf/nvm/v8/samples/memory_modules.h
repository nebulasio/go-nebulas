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

#ifndef _NEBULAS_NF_NVM_V8_LIB_MEMORY_MODULES_H_
#define _NEBULAS_NF_NVM_V8_LIB_MEMORY_MODULES_H_

#include <stddef.h>

char *RequireDelegateFunc(void *handler, const char *filename,
                          size_t *lineOffset);

char *AttachLibVersionDelegateFunc(void *handler, const char *libname);

void AddModule(void *handler, const char *filename, const char *source,
               int lineOffset);

char *GetModuleSource(void *handler, const char *filename);

#endif // _NEBULAS_NF_NVM_V8_LIB_MEMORY_MODULES_H_
