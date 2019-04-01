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
char *AttachLibVersionDelegateFunc(void *handler, const char *libname);

// storage.
char *StorageGetFunc(void *handler, const char *key, size_t *gasCnt);
int StoragePutFunc(void *handler, const char *key, const char *value, size_t *gasCnt);
int StorageDelFunc(void *handler, const char *key, size_t *gasCnt);

// blockchain.
char *GetTxByHashFunc(void *handler, const char *hash, size_t *gasCnt);
int GetAccountStateFunc(void *handler, const char *address, size_t *gasCnt, char **result, char **info);
int TransferFunc(void *handler, const char *to, const char *value, size_t *gasCnt);
int VerifyAddressFunc(void *handler, const char *address, size_t *gasCnt);
int GetPreBlockHashFunc(void *handler, unsigned long long offset, size_t *gasCnt, char **result, char **info);
int GetPreBlockSeedFunc(void *handler, unsigned long long offset, size_t *gasCnt, char **result, char **info);
char *GetContractSourceFunc(void *handler, const char *address, size_t *gasCnt);
char *InnerContractFunc(void *handler, const char *address, const char *funcName, const char * v,
		const char *args, size_t *gasCnt);

//random.
int GetTxRandomFunc(void *handler, size_t *gasCnt, char **result, char **exceptionInfo);
int GetLatestNebulasRankFunc(void *handler, const char *address, size_t *gasCnt, char **result, char **info);
int GetLatestNebulasRankSummaryFunc(void *handler, size_t *gasCnt, char **result, char **info);

// event.
void EventTriggerFunc(void *handler, const char *topic, const char *data, size_t *gasCnt);

// crypto
char *Sha256Func(const char *data, size_t *gasCnt);
char *Sha3256Func(const char *data, size_t *gasCnt);
char *Ripemd160Func(const char *data, size_t *gasCnt);
char *RecoverAddressFunc(int alg, const char *data, const char *sign, size_t *gasCnt);
char *Md5Func(const char *data, size_t *gasCnt);
char *Base64Func(const char *data, size_t *gasCnt);

// The gateway functions.
void V8Log_cgo(int level, const char *msg) {
	V8Log(level, msg);
};

char *RequireDelegateFunc_cgo(void *handler, const char *filename, size_t *lineOffset) {
	return RequireDelegateFunc(handler, filename, lineOffset);
}

char *AttachLibVersionDelegateFunc_cgo(void *handler, const char *libname) {
	return AttachLibVersionDelegateFunc(handler, libname);
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
int GetAccountStateFunc_cgo(void *handler, const char *address, size_t *gasCnt, char **result, char **info) {
	return GetAccountStateFunc(handler, address, gasCnt, result, info);
};
int TransferFunc_cgo(void *handler, const char *to, const char *value, size_t *gasCnt) {
	return TransferFunc(handler, to, value, gasCnt);
};
int VerifyAddressFunc_cgo(void *handler, const char *address, size_t *gasCnt) {
	return VerifyAddressFunc(handler, address, gasCnt);
};

int GetPreBlockHashFunc_cgo(void *handler, unsigned long long offset, size_t *gasCnt, char **result, char **info) {
	return GetPreBlockHashFunc(handler, offset, gasCnt, result, info);
}

int GetPreBlockSeedFunc_cgo(void *handler, unsigned long long offset, size_t *gasCnt, char **result, char **info) {
	return GetPreBlockSeedFunc(handler, offset, gasCnt, result, info);
}

char *GetContractSourceFunc_cgo(void *handler, const char *address, size_t *gasCnt) {
	return GetContractSourceFunc(handler, address, gasCnt);
};
char *InnerContractFunc_cgo(void *handler, const char *address, const char *funcName, const char * v, const char *args, size_t *gasCnt) {
	return InnerContractFunc(handler, address, funcName, v, args, gasCnt);
};
int GetTxRandomFunc_cgo(void *handler, size_t *gasCnt, char **result, char **exceptionInfo) {
	return GetTxRandomFunc(handler, gasCnt, result, exceptionInfo);
};

int GetLatestNebulasRankFunc_cgo(void *handler, const char *address, size_t *gasCnt, char **result, char **info) {
 	return GetLatestNebulasRankFunc(handler, address, gasCnt, result, info);
};
int GetLatestNebulasRankSummaryFunc_cgo(void *handler, size_t *gasCnt, char **result, char **info) {
 	return GetLatestNebulasRankSummaryFunc(handler, gasCnt, result, info);
};

void EventTriggerFunc_cgo(void *handler, const char *topic, const char *data, size_t *gasCnt) {
	EventTriggerFunc(handler, topic, data, gasCnt);
};

char *Sha256Func_cgo(const char *data, size_t *gasCnt) {
	return Sha256Func(data, gasCnt);
}
char *Sha3256Func_cgo(const char *data, size_t *gasCnt) {
	return Sha3256Func(data, gasCnt);
}
char *Ripemd160Func_cgo(const char *data, size_t *gasCnt) {
	return Ripemd160Func(data, gasCnt);
}
char *RecoverAddressFunc_cgo(int alg, const char *data, const char *sign, size_t *gasCnt) {
	return RecoverAddressFunc(alg, data, sign, gasCnt);
}
char *Md5Func_cgo(const char *data, size_t *gasCnt) {
	return Md5Func(data, gasCnt);
}
char *Base64Func_cgo(const char *data, size_t *gasCnt) {
	return Base64Func(data, gasCnt);
}

*/
import "C"
