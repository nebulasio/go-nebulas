
## Overview

## NBRE Architecture

![nbre_construct.jpg](https://github.com/nebulasio/go-nebulas/raw/feature/nbre/nbre/doc/jpg/nbre_construct.jpg)

## NBRE Modules
* common
* fs
* core
* jit
* cmd

## Build NBRE From Scratch
* Set Up Build Environment:
  Go to directory "go-nebulas/nbre", then execute script "prepare.sh"
  This step will compile all the 3rd-party libraries that nbre depends on, and the applications used to compile NBRE source code
  Afterwards, set up environment variables accordingly

```
$> cd go-nebulas/nbre
$> ./prepare.sh
$> source env.set.sh
```

* Compile NBRE source code
  Create directory "build", then generate makefile via cmake
```
$> mkdir build
$> cd build
$> cmake --CMAKE_BUILD_TYPE=Release ../
```

* NBRE Configuration
  * configuration of nasir
```
{
  "name": "nr",
  "version_major": 0,
  "version_minor": 0,
  "version_patch": 1,
  "depends": [
  ],
  "available_height": 100,
  "cpp_files": [
    "test/link_example/foo_arg.cpp"
  ],
  "include_header_files":[
    "test/link_example"
  ],
  "link_path":[
  ],
  "link_files":[
  ],
  "flags": [
  ]
}
```

|Label|Meaning|
|---|:---|
|version_major| major version number |
|version_minor| minor version number |
|version_patch| patch version number |
|depends| dependancies |
|available_height| height that the ir takes effect |
|cpp_files| c++ implementation of core protocols |
|include_header_files| c++ header files of core protocols implementation |
|link_path| linking path(s) |
|link_files| linking file(s) |
|flags| compiling parameters |

* Database Configuration 
  * Edit file "env.set.sh" to add the following line at the end
```
export NBRE_DB=go-nebulas/data.db
```

## Module common
* ipc(inter process communication)
  basic functionalities of neb_ipc, including: bookkeeper, service, session, queue etc.
* util
  * bytes: base58 and base64 encoding and decoding functionalities
  * enable_func_if: templates for checking function types
  * singleton: singleton templates
  * version: version management
* configuration: read ini file and get parameters of JIT
* ir_conf_reader: read json file and get parameters of nasir
* quitable_thread: exception handling and thread exit

## Module fs
* Inheritance relationship between fs classes 

![nbre_fs.jpg](https://github.com/nebulasio/go-nebulas/raw/feature/nbre/nbre/doc/jpg/nbre_fs.jpg)

* rocksdb_storage: rocksdb related operations
* nbre_storage: read/write blockchain related info based on rocksdb_storage
* blockchain: transaction and block data
* protocol buffer

| file | meaning |
|---|:---|
|ir.proto| meta data of ir, including: name, version, effective height, dependancies etc. |
|dag.proto| dag data structure |
|state.proto| state data structure |
|block.proto.patch| block data structure |

* util

## Module core
* Version management
Every 15 seconds, ir_warden will query the block info, and update the ir version number if necessary

* neb_ipc
  Inter process communication between NBRE and Neb

* Check the following flowchat to see the functionalities of version management and ipc


![nbre_ipc.jpg](https://github.com/nebulasio/go-nebulas/blob/feature/nbredev/nbre/doc/jpg/nbre_ipc.jpg)

## Module JIT
NBRE JIT is based on LLVM, core protocol IR is running on it.

* NBRE execution time order


![nbre_jit_sequence.jpg](https://github.com/nebulasio/go-nebulas/raw/feature/nbre/nbre/doc/jpg/nbre_jit_sequence.jpg)

  * jit_driver
    jit_driver serves as the entry of JIT

## Module cmd
  * nasir
    generate ir, more specifically: 

```
Generate IR Payload:
  --help show help message
  --input arg IR configuration file
  --output arg output file
  --mode arg (=payload) Generate ir bitcode or ir payload. - [bitcode |
                        payload], default:payload

```
* payload mode (default mode): carry on the name, version etc, which can be executed directly on JIT.
* bitcode mode: it generates the pure IR code, cannot be executed on JIT, just used for debugging.
