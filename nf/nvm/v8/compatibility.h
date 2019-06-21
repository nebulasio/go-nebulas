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

#include <iostream>
#include <unordered_map>
#include <vector>
#include <string>

const uint32_t MainNetID = 1;
const uint32_t TestNetID = 1001;
const uint32_t LocalNetID = 1111;

const uint64_t DefaultTimeoutValueInMS = 5000000;

namespace SNVM{

    typedef struct version{
        int32_t vmajor;
        int32_t vminor;
        int32_t vpatch;
        version():vmajor(0), vminor(0), vpatch(0){}
    }Version;

    Version* ParseVersion(std::string&);
    int32_t CompareVersion(Version*, Version*);

    class CompatManager{
        public:
            explicit CompatManager(){
                MainNetVersionHeightMap.clear();
                TestNetVersionHeightMap.clear();
                MainNetVersionHeightMap["1.0.5"] = 467500;
                MainNetVersionHeightMap["1.1.0"] = 2188985;
                TestNetVersionHeightMap["1.0.5"] = 424400;
                TestNetVersionHeightMap["1.1.0"] = 600600;
                LocalVersionHeightMap["1.0.5"] = 2;
                LocalVersionHeightMap["1.1.0"] = 2;
                InitializeLibVersionManager();
                InitializeHeightTimeoutMap();
            }
            ~CompatManager(){
                lib_version_manager.clear();
                height_timeout_map.clear();
            }

            void InitializeVersionMap();
            void InitializeHeightTimeoutMap();
            void InitializeLibVersionManager();

            std::string AttachVersionForLib(std::string&, uint64_t, std::string&, uint32_t);
            std::string GetNearestInstructionCounterVersionAtHeight(uint64_t, uint32_t);
            std::string AttachDefaultVersionLib(std::string&);
            std::string FindLastNearestLibVersion(std::string&, std::string&);

            inline uint64_t GetNVMMemoryLimitWithoutInjectHeight(uint32_t curr_chain_id){
                if(curr_chain_id == MainNetID){
                    return MainNetNvmMemoryLimitWithoutInjectHeight;
                }else if(curr_chain_id == TestNetID){
                    return TestNetNvmMemoryLimitWithoutInjectHeight;
                }else{
                    return 0;
                }
            }

            inline uint64_t GetNVMGasLimitWithoutTimeoutHeight(uint32_t curr_chain_id){
                if(curr_chain_id == MainNetID){
                    return MainNetNvmGasLimitWithoutTimeoutAtHeight;
                }else if(curr_chain_id == TestNetID){
                    return TestNetNvmGasLimitWithoutTimeoutAtHeight;
                }else{
                    return 0;
                } 
            }

            inline uint64_t InnerContractCallAvailableHeight(uint32_t curr_chain_id){
                if(curr_chain_id == MainNetID){
                    return MainNetInnerContractCallAvailableAtHeight;
                }else if(curr_chain_id == TestNetID){
                    return TestNetInnerContractCallAvailableAtHeight;
                }else{
                    return 0;
                }
            }

            inline uint64_t V8JSLibVersionControlHeight(uint32_t curr_chain_id){
                if(curr_chain_id == MainNetID)
                    return MainNetV8JSLibVersionControlHeight;
                else if(curr_chain_id == TestNetID)
                    return TestNetV8JSLibVersionControlHeight;
                else
                    return 0L;
            }

            inline uint64_t GetTimeoutConfig(uint64_t block_height, uint32_t chain_id){
                auto item = height_timeout_map.find(block_height);
                if(item != height_timeout_map.end()){
                    auto config = item->second.find(chain_id);
                    if(config != item->second.end()){
                        return config->second;
                    }
                }
                return DefaultTimeoutValueInMS;
            }

        private:
            // NvmMemoryLimitWithoutInjectHeight memory of nvm contract without inject code
            const uint64_t TestNetNvmMemoryLimitWithoutInjectHeight = 281800;
            const uint64_t MainNetNvmMemoryLimitWithoutInjectHeight = 306800;
            //NvmGasLimitWithoutTimeoutAtHeight
            const uint64_t TestNetNvmGasLimitWithoutTimeoutAtHeight = 600000;
            const uint64_t MainNetNvmGasLimitWithoutTimeoutAtHeight = 624763;
            //InnerContractCallAvailableAtHeight
            //const uint64_t TestNetInnerContractCallAvailableAtHeight = 600600;
            //const uint64_t MainNetInnerContractCallAvailableAtHeight = 2188985;
            //TODO: change this before deployment
            const uint64_t TestNetInnerContractCallAvailableAtHeight = 1;
            const uint64_t MainNetInnerContractCallAvailableAtHeight = 1;
            const uint64_t LocalInnerContractCallAvailableAtHeight = 1;

            std::string JSLibRootName = "lib/";
            uint32_t JSLibRootNameLen = 4;
            uint64_t MainNetV8JSLibVersionControlHeight = 467500;
            uint64_t TestNetV8JSLibVersionControlHeight = 424400;
            uint64_t LocalNetV8JSLibVersionControlHeight = 2;

            std::string DefaultV8JSLibVersion = "1.0.0";
            std::vector<std::string> MainNetVersions{"1.1.0", "1.0.5"};
            std::vector<std::string> TestNetVersions{"1.1.0", "1.0.5"};
            std::vector<std::string> LocalVersions{"1.1.0", "1.0.5"};
            std::unordered_map<std::string, uint64_t> MainNetVersionHeightMap;
            std::unordered_map<std::string, uint64_t> TestNetVersionHeightMap;
            std::unordered_map<std::string, uint64_t> LocalVersionHeightMap;

            // data structures
            std::unordered_map<std::string, std::vector<std::string>> lib_version_manager;          // js lib version manager
            std::unordered_map<uint64_t, std::unordered_map<uint32_t, uint64_t>> height_timeout_map;// map<3156680, map<chain_id, 5000000>>
    };
}