
## 概述

## NBRE 结构框架

![nbre_construct.jpg](https://github.com/nebulasio/go-nebulas/raw/feature/nbre/nbre/doc/jpg/nbre_construct.jpg)

## NBRE 功能模块
* common
* fs
* core
* jit
* cmd

## NBRE 开发环境的部署
* 设置环境变量
```

$> cd go-nebulas/nbre
$> source env.set.sh
```
* 编译配置基础环境
  进入NBRE目录，执行prepare.sh 编译基础环境
```
$> cd go-nebulas/nbre
$> ./prepare.sh
```
* 编译NBRE
  建立build目录，执行cmake
```
$> mkdir build
$> cd build
$> cmake ../
```

* 配置NBRE
  *  nasir 的配置
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

|标签名|表述|
|---|:---|
|version_major| 主版本号|
|version_minor| 次版本号|
|version_patch| 修订号|
|depends| 依赖|
|available_height| 有效高度|
|cpp_files| 用于编译IR的cpp文件名称|
|include_header_files|用于编译IR的c++头文件名称|
|link_path|链接路径|
|link_files|链接文件|
|flags|编译器参数|

* 数据库配置
编辑文件env.set.sh
```
export NBRE_DB=go-nebulas/data.db
```
## 模块 common
* ipc
  neb_ipc的基础功能代码，包括bookkeeper、service、session、queue等内容
* util
  * bytes类拥有base58、base64的解码和译码功能
  * enable_func_if 负责判断函数类型的模板
  * singleton 实现单例的基类模板
  * version 版本类
* configuration 读取ini文件，负责获得jit所需的参数
* ir_conf_reader 读取json文件，负责获取nasir所需的参数
* quitable_thread 能够妥善处理异常后，安全退出线程的基类
## 模块 fs
* fs类之间的关系图

![nbre_fs.jpg](https://github.com/nebulasio/go-nebulas/raw/feature/nbre/nbre/doc/jpg/nbre_fs.jpg)

* rocksdb_storage 负责rocksdb数据库的增删改查操作
* nbre_storage 依赖rocksdb_storage进行读写blockchain
* blockchain 区块信息类，依赖protocol生成的Block对象
* protocol buffer

|文件|表述|
|---|:---|
|ir.proto| ir 名称、版本、有效高度、依赖等元数据表述 |
|dag.proto|dag数据表述 |
|state.proto|状态数据表述 |
|block.proto.patch|区块数据表述 |

* util

## 模块 core
* 版本控制
ir_warden每15秒钟一次，查询区块，并更新版本号

* neb_ipc
  主要负责nbre与neb之间的通讯

* 通过以下流程图说明版本控制和ipc的作用


![nbre_ir_warden_flow.jpg](https://github.com/nebulasio/go-nebulas/raw/feature/nbre/nbre/doc/jpg/nbre_ir_warden_flow.jpg)

## 模块 jit
nbre jit通过llvm的JIT(Just-in-time compilation),编译并执行ir

* nbre执行时序


![nbre_jit_sequence.jpg](https://github.com/nebulasio/go-nebulas/raw/feature/nbre/nbre/doc/jpg/nbre_jit_sequence.jpg)

  * jit_driver
    jit_driver是jit模块的入口程序
## 模块  cmd
  * nasir
    负责生成ir, 具体功能如下：

```
Generate IR Payload:
  --help show help message
  --input arg IR configuration file
  --output arg output file
  --mode arg (=payload) Generate ir bitcode or ir payload. - [bitcode |
                        payload], default:payload

```
* payload模式为默认状态，携带名称、版本号等相关描述信息，可以由nbre jit执行的二进制代码
* bitcode模式将输出纯的ir，nbre jit不能直接执行，用作clang调试使用














