// Copyright (C) 2017-2019 go-nebulas authors
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
// Author: Samuel Chen <samuel.chen@nebulas.io>

#pragma once

#define REQUIRE_DELEGATE_FUNC               "RequireDelegateFunc"
#define ATTACH_LIB_VERSION_DELEGATE_FUNC    "AttachLibVersionDelegateFunc"
#define STORAGE_GET                         "StorageGet"
#define STORAGE_PUT                         "StoragePut"
#define STORAGE_DEL                         "StorageDel"
#define GET_TX_BY_HASH                      "GetTxByHash"
#define GET_ACCOUNT_STATE                   "GetAccountState"
#define TRANSFER                            "Transfer"
#define VERIFY_ADDR                         "VerifyAddress"
#define GET_PRE_BLOCK_HASH                  "GetPreBlockHash"
#define GET_PRE_BLOCK_SEED                  "GetPreBlockSeed"
#define EVENT_TRIGGER_FUNC                  "EventTriggerFunc"
#define SHA_256_FUNC                        "Sha256Func"
#define SHA_3256_FUNC                       "Sha3256Func"
#define RIPEMD_160_FUNC                     "Ripemd160Func"
#define RECOVER_ADDRESS_FUNC                "RecoverAddressFunc"
#define MD5_FUNC                            "Md5Func"
#define BASE64_FUNC                         "Base64Func"

// inner contract call
#define GET_CONTRACT_SRC                    "GetContractSource"
#define INNER_CONTRACT_CALL                 "InnerContractCall"

// nr
#define GET_LATEST_NR                       "GetLatestNebulasRank"
#define GET_LATEST_NR_SUMMARY               "GetLatestNebulasRankSummary"