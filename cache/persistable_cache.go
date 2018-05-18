package cache

import (
	"encoding/gob"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

type replay struct {
	op Operator
	k  interface{}
	v  interface{}
}

// PersistableCache ..
type PersistableCache struct {
	shadowCache map[interface{}]interface{}
	originCache Cache
	conf        *PersistableCacheConfig
	incFile     *os.File
	incEncoder  *gob.Encoder
	replayChan  chan *replay
}

// NewPersistableCache ..
func NewPersistableCache(cache Cache, conf *PersistableCacheConfig) Cache {
	pc := &PersistableCache{
		shadowCache: make(map[interface{}]interface{}),
		originCache: cache,
		conf:        conf,
		replayChan:  make(chan *replay, conf.CacheSize),
	}
	pc.start()
	return pc
}

// Entries ..
func (pc *PersistableCache) Entries() interface{} {
	return pc.originCache.Entries()
}

// Keys ..
func (pc *PersistableCache) Keys() interface{} {
	return pc.originCache.Keys()
}

// Set ...
func (pc *PersistableCache) Set(k, v interface{}) (err error) {

	pc.replayChan <- &replay{
		Add,
		k,
		v,
	}
	return pc.originCache.Set(k, v)
}

// Get ...
func (pc *PersistableCache) Get(k interface{}) interface{} {
	return pc.originCache.Get(k)
}

// Size ..
func (pc *PersistableCache) Size() int {
	return pc.originCache.Size()
}

// Take ..
func (pc *PersistableCache) Take(k interface{}) (interface{}, error) {
	pc.replayChan <- &replay{
		op: Delete,
		k:  k,
	}
	return pc.originCache.Take(k)
}

// Encode ..
func (pc *PersistableCache) Encode(k, v interface{}) (*ExportableEntry, error) {
	return pc.originCache.Encode(k, v)
}

// Decode ..
func (pc *PersistableCache) Decode(kv *ExportableEntry) (interface{}, interface{}, error) {
	return pc.originCache.Decode(kv)
}

func (pc *PersistableCache) start() {

	fi, err := os.Stat(pc.conf.SaveDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(pc.conf.SaveDir, os.ModeDir|os.ModePerm)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
				"dir": pc.conf.SaveDir,
			}).Fatal("Create save dir failed.")
		}
	} else {
		if !fi.IsDir() {
			logging.VLog().WithFields(logrus.Fields{
				"name": pc.conf.SaveDir,
			}).Fatal("File already exists with the same name.")
		}
	}

	ticker := time.NewTicker(time.Millisecond * time.Duration(pc.conf.IntervalMs))
	pc.newLogger(strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond))))

	go func() {
		last := 0
		for r := range pc.replayChan {
			select {
			case <-ticker.C:
				start := int(time.Now().UnixNano() / int64(time.Millisecond))
				now := strconv.Itoa(start)
				pc.newLogger(now)
				pc.dump(now, strconv.Itoa(last))
				end := int(time.Now().UnixNano() / int64(time.Millisecond))

				logging.VLog().WithFields(logrus.Fields{
					"cost":      end - start,
					"dir":       pc.conf.SaveDir,
					"timestamp": start,
				}).Info("Cache dump done.")

				last = start
			default:
				// non-block
			}

			// TODO: check nil
			var l *SerializableLog
			// replay
			switch r.op {
			case Add:
				if _, ok := pc.shadowCache[r.k]; !ok {
					pc.shadowCache[r.k] = r.v
					kv, err := pc.Encode(r.k, r.v)
					if err != nil {
						logging.VLog().WithFields(logrus.Fields{
							"err": err,
						}).Error("Encode ExportableEntry error.")
					} else {
						l = &SerializableLog{
							Add,
							kv,
						}
					}
				}
			case Delete:
				delete(pc.shadowCache, r.k)
				kv, err := pc.Encode(r.k, nil)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err": err,
					}).Error("Encode ExportableEntry error.")
				} else {
					l = &SerializableLog{
						Delete,
						kv,
					}
				}
			default:
				logging.VLog().WithFields(logrus.Fields{
					"type": r.op,
				}).Error("Unknown op type.")
			}

			// inc log
			if l != nil {
				err := pc.incEncoder.Encode(l)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err":  err,
						"item": l, // TODO: implements String()
					}).Error("Recording log error.")
				}
			}
		}
	}()
}

func (pc *PersistableCache) dump(now, last string) {

	// dump snapshot
	f, err := os.Create(path.Join(pc.conf.SaveDir, now+".snap"))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":       err,
			"timestamp": now,
		}).Fatal("Create snapshot file error.") // TODO: Fatal?
	}
	snapEncoder := gob.NewEncoder(f)

	for k, v := range pc.shadowCache {
		kv, err := pc.Encode(k, v)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Encode ExportableEntry error.")
			continue
		}
		err = snapEncoder.Encode(kv)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
				"key": k,
			}).Error("Write snapshot item error.")
		}
	}

	err = f.Close()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":  err,
			"file": f.Name(),
		}).Error("Close snapshot file error.")
	}

	// create done file
	doneF, err := os.Create(path.Join(pc.conf.SaveDir, now+".done"))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":       err,
			"timestamp": now,
		}).Fatal("Create done file error.") // TODO: Fatal?
	} else {
		defer doneF.Close()
	}

	// delete last files
	lastFiles := []string{
		path.Join(pc.conf.SaveDir, last+".snap"),
		path.Join(pc.conf.SaveDir, last+".done"),
		path.Join(pc.conf.SaveDir, last+".inc"),
	}
	for _, f := range lastFiles {
		_, err = os.Stat(f)
		if os.IsNotExist(err) {
			continue
		}

		err = os.Remove(f)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":      err,
				"filename": f,
			}).Error("Delete old file error.")
		}
	}
}

func (pc *PersistableCache) newLogger(now string) {
	if pc.incFile != nil {
		err := pc.incFile.Close()
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"file": pc.incFile.Name(),
			}).Error("Close last inc log file error.")
		}
	}

	incFile, err := os.Create(filepath.Join(pc.conf.SaveDir, now+".inc"))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":      err,
			"filename": incFile.Name(),
		}).Fatal("Create inc log file error.") // TODO: Fatal?
	}

	pc.incFile = incFile
	pc.incEncoder = gob.NewEncoder(incFile)

	logging.VLog().WithFields(logrus.Fields{
		"file": pc.incFile.Name(),
	}).Info("Create new inc log file.")
}
