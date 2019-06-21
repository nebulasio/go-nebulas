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

#include "nvm_engine.h"
#include "compatibility.h"

SNVM::Version* SNVM::ParseVersion(std::string& version_str){
    if(version_str.length() == 0)
        return nullptr;

    try{
        std::vector<std::string> str_vec;
        std::string buf;
        char delimer = '.';
        uint64_t i=0;
        while(i<version_str.length()){
            if(version_str[i] != delimer){
                buf += version_str[i];
            }else if(buf.length()>0){
                str_vec.push_back(buf);
                buf = "";
            }
            i++;
        }
        if(!buf.empty()){
            str_vec.push_back(buf);
        }

        if(str_vec.size() != 3)
            return nullptr;

        int32_t major = (int32_t)std::stoi(str_vec[0]);
        if(major < 0)
            return nullptr;

        int32_t minor = (int32_t)std::stoi(str_vec[1]);
        if(minor < 0)
            return nullptr;

        int32_t patch = (int32_t)std::stoi(str_vec[2]);
        if(patch < 0)
            return nullptr;

        SNVM::Version* parsed_version = new SNVM::Version();
        parsed_version->vmajor = major;
        parsed_version->vminor = minor;
        parsed_version->vpatch = patch;

        return parsed_version;

    }catch(const std::exception& e){
        if(FG_DEBUG)
        std::cout<<e.what()<<std::endl;
    }
    return nullptr;
}

int32_t SNVM::CompareVersion(Version* a, Version* b){
    if(a->vmajor > b->vmajor){
        return 1;
    }
    if(a->vmajor < b->vmajor){
        return -1;
    }
    if(a->vminor > b->vminor){
        return 1;
    }
    if(a->vminor < b->vminor){
        return -1;
    }
    if(a->vpatch > b->vpatch){
        return 1;
    }
    if(a->vpatch < b->vpatch){
        return -1;
    }
    return 0;
}

void SNVM::CompatManager::InitializeHeightTimeoutMap(){
    height_timeout_map.insert(std::pair<uint64_t, 
            std::unordered_map<uint32_t, uint64_t>>(233584, std::unordered_map<uint32_t, uint64_t>{{MainNetID, 5000100}}));
}

void SNVM::CompatManager::InitializeLibVersionManager(){

    std::string default_version("1.0.0");
    try{
        lib_version_manager.clear();
        
        std::vector<std::string> vec_execution_env{"1.0.0", "1.0.5"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("execution_env.js", vec_execution_env));

        std::vector<std::string> vec_bignumber{"1.0.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("bignumber.js", vec_bignumber));

        std::vector<std::string> vec_random{"1.0.0", "1.0.5", "1.1.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("random.js", vec_random));

        std::vector<std::string> vec_date{"1.0.0", "1.0.5"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("date.js", vec_date));
        
        std::vector<std::string> vec_tsc{"1.0.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("tsc.js", vec_tsc));

        std::vector<std::string> vec_util{"1.0.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("util.js", vec_util));

        std::vector<std::string> vec_esprima{"1.0.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("esprima.js", vec_esprima));

        std::vector<std::string> vec_assert{"1.0.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("assert.js", vec_assert));

        std::vector<std::string> vec_inst_counter{"1.0.0", "1.1.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("instruction_counter.js", vec_inst_counter));

        std::vector<std::string> vec_type_script{"1.0.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("typescriptServices.js", vec_type_script));

        std::vector<std::string> vec_blockchain{"1.0.0", "1.0.5", "1.1.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("blockchain.js", vec_blockchain));

        std::vector<std::string> vec_console{"1.0.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("console.js", vec_console));

        std::vector<std::string> vec_event{"1.0.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("event.js", vec_event));

        std::vector<std::string> vec_storage{"1.0.0"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("storage.js", vec_storage));

        std::vector<std::string> vec_crypto{"1.0.5"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("crypto.js", vec_crypto));

        std::vector<std::string> vec_uint{"1.0.5"};
        lib_version_manager.insert(std::pair<std::string, std::vector<std::string>>("uint.js", vec_uint));

    }catch(const std::exception& e){
        if(FG_DEBUG)
            std::cout<<e.what()<<std::endl;
    }
}

std::string SNVM::CompatManager::GetNearestInstructionCounterVersionAtHeight(uint64_t block_height, uint32_t curr_chain_id) {    
    if(curr_chain_id == MainNetID){
        for(auto iter=this->MainNetVersions.begin(); iter!=this->MainNetVersions.end(); iter++){
            if(block_height >= this->MainNetVersionHeightMap[*iter])
                return *iter;
        }
    }else if(curr_chain_id == TestNetID){
        for(auto iter=this->TestNetVersions.begin(); iter!=this->TestNetVersions.end(); iter++){
            if(block_height >= this->TestNetVersionHeightMap[*iter])
                return *iter;
        }
    }else{
        for(auto iter=this->LocalVersions.begin(); iter!=this->LocalVersions.end(); iter++){
            if(block_height >= this->LocalVersionHeightMap[*iter])
                return *iter;
        }
    }

	return std::string("1.0.0");
}

std::string SNVM::CompatManager::FindLastNearestLibVersion(std::string& deploy_version_str, std::string& lib_file_name){
    if(deploy_version_str.length() == 0 || lib_file_name.length() == 0){
        LogErrorf("FindLastNearestLibVersion: empty arguments");
        return "";
    }

    try{
        Version* deploy_version = ParseVersion(deploy_version_str);
        auto vec_iter = lib_version_manager.find(lib_file_name);
        if( vec_iter != lib_version_manager.end()){
            std::vector<std::string> version_vec = vec_iter->second;
            for(auto iter=version_vec.rbegin(); iter!=version_vec.rend(); ++iter){
                Version* lib_version = ParseVersion(*iter);
                if(CompareVersion(lib_version, deploy_version) <= 0){
                    LogDebugf("Find version %s for lib %s with given deploy version %s", 
                            static_cast<const char*>((*iter).c_str()), 
                            static_cast<const char*>(lib_file_name.c_str()), 
                            static_cast<const char*>(deploy_version_str.c_str()));
                    delete deploy_version;
                    return *iter;
                }
            }
        }else{
            LogDebugf("js lib not configured, libname: %s, deployversion: %s", 
                    static_cast<const char*>(lib_file_name.c_str()), 
                    static_cast<const char*>(deploy_version_str.c_str()));
        }
        if(deploy_version != nullptr)
            delete deploy_version;

    }catch(const std::exception &e){
        if(FG_DEBUG)
            std::cout<<e.what()<<std::endl;
        std::cout<<e.what()<<std::endl;
    }
	return "";
}


std::string SNVM::CompatManager::AttachVersionForLib(
    std::string& lib_name, 
    uint64_t block_height, 
    std::string& meta_version,
    uint32_t chain_id){

    std::string empty_res("");

    // lib/execution_env.js
    std::string inst_counter_file_name("instruction_counter.js");

    // if lib_name endwith "instruction_counter.js"
    if(lib_name.length()>=inst_counter_file_name.length() &&  (lib_name.compare(lib_name.length()-inst_counter_file_name.length(), 
            inst_counter_file_name.length(), inst_counter_file_name) == 0)){

        std::string version = this->GetNearestInstructionCounterVersionAtHeight(block_height, chain_id);
        std::string version_file_path = this->JSLibRootName + version 
                + lib_name.substr(this->JSLibRootNameLen-1, lib_name.length()-this->JSLibRootNameLen+1);
        return version_file_path;
    }

	// block after core.V8JSLibVersionControlHeight, inclusive
	if(block_height >= this->V8JSLibVersionControlHeight(chain_id)){

        if(meta_version.length() == 0){
            LogDebugf("Contract meta is nil for %s at height: %lld", static_cast<const char*>(lib_name.c_str()), block_height);
			return this->AttachDefaultVersionLib(lib_name);
		}
        if(lib_name.compare(0, this->JSLibRootName.length(), this->JSLibRootName)!=0 ||
            lib_name.find("../") != std::string::npos){
            return empty_res;
        }

        // lib/inst.js
        std::string lib_file_name = lib_name.substr(this->JSLibRootNameLen, lib_name.length()-this->JSLibRootNameLen);
		std::string ver = this->FindLastNearestLibVersion(meta_version, lib_file_name);
        if(ver.length() == 0){
            LogErrorf("lib version not found. libname: %s, deployLibVer: %s", 
                    static_cast<const char*>(lib_name.c_str()), 
                    static_cast<const char*>(meta_version.c_str()));
            return empty_res;
        }

        // e.g: lib/1.0.0/date.js
		return JSLibRootName + ver + lib_name.substr(JSLibRootNameLen-1, lib_name.length()-JSLibRootNameLen+1);
	}

	return this->AttachDefaultVersionLib(lib_name);
}

std::string SNVM::CompatManager::AttachDefaultVersionLib(std::string& lib_name){
    // if lib_name startswith "lib/"
    if(lib_name.compare(0, JSLibRootNameLen, JSLibRootName) != 0){
        if(lib_name[0] == '/'){
            lib_name = "lib" + lib_name;
        }else{
            lib_name = JSLibRootName + lib_name;
        }
    }
    return JSLibRootName + DefaultV8JSLibVersion + lib_name.substr(JSLibRootNameLen-1, lib_name.length()-JSLibRootNameLen+1);
}