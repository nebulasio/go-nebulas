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
#include <unordered_map>
#include <vector>
#include <string>

const uint32_t MainNetID = 1;
const uint32_t TestNetID = 1001;

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
            explicit CompatManager(uint32_t chain_id){
                curr_chain_id = chain_id;
                std::cout<<"**** curr chain id: "<<curr_chain_id<<std::endl;
                MainNetVersionHeightMap.clear();
                TestNetVersionHeightMap.clear();
                MainNetVersionHeightMap["1.0.5"] = 467500;
                MainNetVersionHeightMap["1.1.0"] = 2188985;
                TestNetVersionHeightMap["1.0.5"] = 424400;
                TestNetVersionHeightMap["1.1.0"] = 600600;
                LocalVersionHeightMap["1.0.5"] = 2;
                LocalVersionHeightMap["1.1.0"] = 2;
                InitializeLibVersionManager();
            }
            ~CompatManager(){
                lib_version_manager.clear();
                version_map.clear();
            }

            void InitializeLibVersionManager();

            std::string AttachVersionForLib(std::string&, uint64_t, std::string&);

            std::string GetNearestInstructionCounterVersionAtHeight(uint64_t);

            std::string AttachDefaultVersionLib(std::string&);

            std::string FindLastNearestLibVersion(std::string&, std::string&);
           

            inline uint64_t GetNVMMemoryLimitWithoutInjectHeight(){
                if(this->curr_chain_id == MainNetID){
                    return MainNetNvmMemoryLimitWithoutInjectHeight;
                }else if(this->curr_chain_id == TestNetID){
                    return TestNetNvmMemoryLimitWithoutInjectHeight;
                }else{
                    return 0;
                }
            }

            inline uint64_t GetNVMGasLimitWithoutTimeoutHeight(){
                if(this->curr_chain_id == MainNetID){
                    return MainNetNvmGasLimitWithoutTimeoutAtHeight;
                }else if(this->curr_chain_id == TestNetID){
                    return TestNetNvmGasLimitWithoutTimeoutAtHeight;
                }else{
                    return 0;
                } 
            }

            inline uint64_t InnerContractCallAvailableHeight(){
                if(this->curr_chain_id == MainNetID){
                    return MainNetInnerContractCallAvailableAtHeight;
                }else if(this->curr_chain_id == TestNetID){
                    return TestNetInnerContractCallAvailableAtHeight;
                }else{
                    return 0;
                }
            }

            inline uint64_t V8JSLibVersionControlHeight(){
                if(this->curr_chain_id == MainNetID)
                    return MainNetV8JSLibVersionControlHeight;
                else if(this->curr_chain_id == TestNetID)
                    return TestNetV8JSLibVersionControlHeight;
                else
                    return 0L;
            }

            inline void SetChainID(uint32_t chain_id){
                curr_chain_id = chain_id;
            }

            inline uint32_t GetChainID(){ return curr_chain_id; }

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

            uint32_t curr_chain_id;

            // data structures
            std::unordered_map<std::string, std::vector<std::string>> lib_version_manager;          // js lib version manager
            std::unordered_map<uint32_t, std::unordered_map<std::string, uint64_t>> version_map;    // map<chain_id, map<"1.1.0", 123456>>
    };
}