// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package nvm

/*
#include <stddef.h>

// logger.
void V8Log(int level, const char *msg);

// require.
char *RequireDelegateFunc(void *handler, const char *filename, size_t *lineOffset);

// storage.
char *StorageGetFunc(void *handler, const char *key);
int StoragePutFunc(void *handler, const char *key, const char *value);
int StorageDelFunc(void *handler, const char *key);

// blockchain.
char *GetTxByHashFunc(void *handler, const char *hash);
char *GetAccountStateFunc(void *handler, const char *address);
int TransferFunc(void *handler, const char *to, const char *value);
int VerifyAddressFunc(void *handler, const char *address);

// event.
void EventTriggerFunc(void *handler, const char *topic, const char *data);

// The gateway functions.
void V8Log_cgo(int level, const char *msg) {
	V8Log(level, msg);
};

char *RequireDelegateFunc_cgo(void *handler, const char *filename, size_t *lineOffset) {
	return RequireDelegateFunc(handler, filename, lineOffset);
}

char *StorageGetFunc_cgo(void *handler, const char *key) {
	return StorageGetFunc(handler, key);
};
int StoragePutFunc_cgo(void *handler, const char *key, const char *value) {
	return StoragePutFunc(handler, key, value);
};
int StorageDelFunc_cgo(void *handler, const char *key) {
	return StorageDelFunc(handler, key);
};

char *GetTxByHashFunc_cgo(void *handler, const char *hash) {
	return GetTxByHashFunc(handler, hash);
};
char *GetAccountStateFunc_cgo(void *handler, const char *address) {
	return GetAccountStateFunc(handler, address);
};
int TransferFunc_cgo(void *handler, const char *to, const char *value) {
	return TransferFunc(handler, to, value);
};
int VerifyAddressFunc_cgo(void *handler, const char *address) {
	return VerifyAddressFunc(handler, address);
};

void EventTriggerFunc_cgo(void *handler, const char *topic, const char *data) {
	EventTriggerFunc(handler, topic, data);
};

*/
import "C"
