## CHANGLOGS

### 2018-04-21

* 添加了 docker-compose 可以跑 seed 和 miner

* 把包含go1.9.2 和 rocksdb5.13 整合在 docker image 里

* 添加了一些功能：
如果在当前目录下 存在vendor 文件夹 则不执行vendor更新；
如果直接下载vendor压缩包，只要在当前目录下添加空文件 nodep 即可
```bash
    touch ./nodep
```
