## CHANGLOGS

### 2018-04-21

* Added docker-compose.yml to run node directly

* Included go1.9.2 and rocksdb5.13 in Dockerfile

* New Feature：
  If you want to download vender instead of running `make dep`,
  you can create `nodep` file with `touch nodep` in root directory
  and then run `docker-compose up node`
