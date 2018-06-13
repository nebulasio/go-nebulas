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
char *StorageGetFunc(void *handler, const char *key, size_t *gasCnt);
int StoragePutFunc(void *handler, const char *key, const char *value, size_t *gasCnt);
int StorageDelFunc(void *handler, const char *key, size_t *gasCnt);

// blockchain.
char *GetTxByHashFunc(void *handler, const char *hash, size_t *gasCnt);
char *GetAccountStateFunc(void *handler, const char *address, size_t *gasCnt);
int TransferFunc(void *handler, const char *to, const char *value, size_t *gasCnt);
int VerifyAddressFunc(void *handler, const char *address, size_t *gasCnt);
char *GetContractSourceFunc(void *handler, const char *address, size_t *gasCnt);
char *InnerContractFunc(void *handler, const char *address, const char *funcName, const char * v,
		const char *args, size_t *gasCnt);

//random.
char *GetTxRandomFunc(void *handler);

// event.
void EventTriggerFunc(void *handler, const char *topic, const char *data, size_t *gasCnt);

// The gateway functions.
void V8Log_cgo(int level, const char *msg) {
	V8Log(level, msg);
};

char *RequireDelegateFunc_cgo(void *handler, const char *filename, size_t *lineOffset) {
	return RequireDelegateFunc(handler, filename, lineOffset);
}

char *StorageGetFunc_cgo(void *handler, const char *key, size_t *gasCnt) {
	return StorageGetFunc(handler, key, gasCnt);
};
int StoragePutFunc_cgo(void *handler, const char *key, const char *value, size_t *gasCnt) {
	return StoragePutFunc(handler, key, value, gasCnt);
};
int StorageDelFunc_cgo(void *handler, const char *key, size_t *gasCnt) {
	return StorageDelFunc(handler, key, gasCnt);
};

char *GetTxByHashFunc_cgo(void *handler, const char *hash, size_t *gasCnt) {
	return GetTxByHashFunc(handler, hash, gasCnt);
};
char *GetAccountStateFunc_cgo(void *handler, const char *address, size_t *gasCnt) {
	return GetAccountStateFunc(handler, address, gasCnt);
};
int TransferFunc_cgo(void *handler, const char *to, const char *value, size_t *gasCnt) {
	return TransferFunc(handler, to, value, gasCnt);
};
int VerifyAddressFunc_cgo(void *handler, const char *address, size_t *gasCnt) {
	return VerifyAddressFunc(handler, address, gasCnt);
};
char *GetContractSourceFunc_cgo(void *handler, const char *address, size_t *gasCnt) {
	return GetContractSourceFunc(handler, address, gasCnt);
};
char *InnerContractFunc_cgo(void *handler, const char *address, const char *funcName, const char * v, const char *args, size_t *gasCnt) {
	return InnerContractFunc(handler, address, funcName, v, args, gasCnt);
};
char *GetTxRandomFunc_cgo(void *handler) {
	return GetTxRandomFunc(handler);
};
void EventTriggerFunc_cgo(void *handler, const char *topic, const char *data, size_t *gasCnt) {
	EventTriggerFunc(handler, topic, data, gasCnt);
};

*/
import "C"
