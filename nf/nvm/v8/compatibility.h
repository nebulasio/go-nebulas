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

#pragma once

#include <iostream>

/********* height compatibility settings for testnet **********/
uint32_t CurrChainID = 1;

const uint32_t MainNetID = 1;
const uint32_t TestNetID = 1001;

// NvmMemoryLimitWithoutInjectHeight memory of nvm contract without inject code
const uint64_t TestNetNvmMemoryLimitWithoutInjectHeight = 281800;
const uint64_t MainNetNvmMemoryLimitWithoutInjectHeight = 306800;

//NvmGasLimitWithoutTimeoutAtHeight
const uint64_t TestNetNvmGasLimitWithoutTimeoutAtHeight = 600000;
const uint64_t MainNetNvmGasLimitWithoutTimeoutAtHeight = 624763;


inline uint64_t GetNVMMemoryLimitWithoutInjectHeight(){
    if(CurrChainID == MainNetID){
        return MainNetNvmMemoryLimitWithoutInjectHeight;
    }else if(CurrChainID == TestNetID){
        return TestNetNvmMemoryLimitWithoutInjectHeight;
    }else{
        return 0;
    }
}

inline uint64_t GetNVMGasLimitWithoutTimeoutAtHeight(){
    if(CurrChainID == MainNetID){
        return MainNetNvmGasLimitWithoutTimeoutAtHeight;
    }else if(CurrChainID == TestNetID){
        return TestNetNvmGasLimitWithoutTimeoutAtHeight;
    }else{
        return 0;
    } 
}