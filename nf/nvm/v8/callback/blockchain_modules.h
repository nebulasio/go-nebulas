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

#pragma once

#include <stddef.h>
#include "../v8_util.h"
#include "callback_util.h"

char *GetTxByHash(void *handler, const char *hash, size_t *gasCnt);
int GetAccountState(void *handler, const char *addres, size_t *gasCnts, char **result, char **info);
int Transfer(void *handler, const char *to, const char *value, size_t *gasCnt);
int VerifyAddress(void *handler, const char *address, size_t *gasCnt);
int GetPreBlockHash(void *handler, unsigned long long offset, size_t *counterVal, char **result, char **info);
int GetPreBlockSeed(void *handler, unsigned long long offset, size_t *counterVal, char **result, char **info);
char *GetContractSourceFunc(void *handler, const char *address, size_t *counterVal);
char *InnerContractFunc(void *handler, const char *address, const char *funcName, const char *v, const char *args, size_t *gasCnt);