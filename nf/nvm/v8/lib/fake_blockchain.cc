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

#include "fake_blockchain.h"

#include <stdio.h>
#include <stdlib.h>
#include <string>

#include <string.h>
#include "nvm_error.h"

using namespace std;

char *GetTxByHash(void *handler, const char *hash, size_t *gasCnt) {
  *gasCnt = 1000;

  char *ret = NULL;
  string value = "{\"hash\":\"5e6d587f26121f96a07cf4b8b569aac1\",\"from\":"
                 "\"70e30fcae5e7f4b2460faaa9e5b1bd912332ebb5\",\"to\":"
                 "\"70e30fcae5e7f4b2460faaa9e5b1bd912332ebb5\",\"value\":1,"
                 "\"nonce\":4}";
  ret = (char *)calloc(value.length() + 1, sizeof(char));
  strncpy(ret, value.c_str(), value.length());
  return ret;
}

int GetAccountState(void *handler, const char *address, size_t *gasCnt, char **result, char **info) {
  *gasCnt = 1000;

  string value = "{\"value\":1,\"nonce\":4}";
  *result = (char *)calloc(value.length() + 1, sizeof(char));
  strncpy(*result, value.c_str(), value.length());
  return NVM_SUCCESS;
}

int Transfer(void *handler, const char *to, const char *value, size_t *gasCnt) {
  *gasCnt = 2000;
  return NVM_SUCCESS;
}

int VerifyAddress(void *handler, const char *address, size_t *gasCnt) {
  *gasCnt = 100;
  return NVM_SUCCESS;
}

int GetPreBlockHash(void *handler, unsigned long long offset, size_t *gasCnt, char **result, char **info) {
  *gasCnt = 1000;
  return NVM_SUCCESS;
}

int GetPreBlockSeed(void *handler, unsigned long long offset, size_t *gasCnt, char **result, char **info) {
  *gasCnt = 1000;
  return NVM_SUCCESS;
}

int GetLatestNebulasRank(void *handler, const char *address, size_t *gasCnt, char **result, char **info) {
  *gasCnt = 20000;
  return NVM_SUCCESS;
}

int GetLatestNebulasRankSummary(void *handler, size_t *gasCnt, char **result, char **info) {
  *gasCnt = 20000;
  return NVM_SUCCESS;
}

