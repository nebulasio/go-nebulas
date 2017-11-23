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

using namespace std;

char *GetTxByHash(void *handler, const char *hash) {
  char *ret = NULL;
  string value = "{\"hash\":\"5e6d587f26121f96a07cf4b8b569aac1\",\"from\":"
                 "\"70e30fcae5e7f4b2460faaa9e5b1bd912332ebb5\",\"to\":"
                 "\"70e30fcae5e7f4b2460faaa9e5b1bd912332ebb5\",\"value\":1,"
                 "\"nonce\":4}";
  ret = (char *)calloc(value.length() + 1, sizeof(char));
  strncpy(ret, value.c_str(), value.length());
  return ret;
}

char *GetAccountState(void *handler, const char *address) {
  char *ret = NULL;
  string value = "{\"value\":1,\"nonce\":4}";
  ret = (char *)calloc(value.length() + 1, sizeof(char));
  strncpy(ret, value.c_str(), value.length());
  return ret;
}

int Transfer(void *handler, const char *to, const char *value) { return 1; }

int VerifyAddress(void *handler, const char *address) { return 1; }
