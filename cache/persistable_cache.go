package cache

import (
	"encoding/gob"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
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
	shadowCache        map[interface{}]interface{}
	originCache        Cache
	conf               *PersistableCacheConfig
	incFile            *os.File
	incEncoder         *gob.Encoder
	replayChan         chan *replay
	persistenceStarted bool
	ticker             *time.Ticker
}

// NewPersistableCache ..
func NewPersistableCache(cache Cache, conf *PersistableCacheConfig) Cache {
	pc := &PersistableCache{
		shadowCache:        make(map[interface{}]interface{}),
		originCache:        cache,
		conf:               conf,
		replayChan:         make(chan *replay, conf.CacheSize),
		persistenceStarted: false,
	}
	pc.init()
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
	err = pc.originCache.Set(k, v)
	if err == nil {
		pc.replayChan <- &replay{
			Add,
			k,
			v,
		}
	}
	return
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
func (pc *PersistableCache) Take(k interface{}) (v interface{}, err error) {
	v, err = pc.originCache.Take(k)
	if err == nil {
		pc.replayChan <- &replay{
			op: Delete,
			k:  k,
		}
	}
	return
}

// Encode ..
func (pc *PersistableCache) Encode(k, v interface{}) (*ExportableEntry, error) {
	return pc.originCache.Encode(k, v)
}

// Decode ..
func (pc *PersistableCache) Decode(kv *ExportableEntry) (interface{}, interface{}, error) {
	return pc.originCache.Decode(kv)
}

// StartPersistence ..
func (pc *PersistableCache) StartPersistence() {
	if pc.persistenceStarted {
		return
	}
	// dump immediately
	now := strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond)))
	pc.changeReplayLogger(now)
	pc.dump(now, "")
	pc.ticker = time.NewTicker(time.Millisecond * time.Duration(pc.conf.IntervalMs))

	pc.persistenceStarted = true
}

func (pc *PersistableCache) init() {

	// ticker := time.NewTicker(time.Millisecond * time.Duration(pc.conf.IntervalMs))
	// pc.newLogger(strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond))))

	pc.restore()

	go func() {
		last := 0
		for r := range pc.replayChan {

			b := pc.replayToShadow(r)

			if !pc.persistenceStarted {
				continue
			}

			if b {
				// records only when replay succeed
				pc.recordReplay(r)
			}

			select {
			case <-pc.ticker.C:
				start := int(time.Now().UnixNano() / int64(time.Millisecond))
				now := strconv.Itoa(start)
				pc.changeReplayLogger(now)
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
		}
	}()
}

func (pc *PersistableCache) restore() {
	// check and mkdirs
	fi, err := os.Stat(pc.conf.SaveDir) // TODO: path
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

	// list files
	files, err := ioutil.ReadDir(pc.conf.SaveDir)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"dir": pc.conf.SaveDir,
		}).Fatal("List files error.")
	}

	// filter max timestamp
	lastest := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasSuffix(f.Name(), ".done") {
			cur, err := strconv.Atoi(strings.Split(f.Name(), ".done")[0])
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err":      err,
					"filename": f.Name(),
				}).Error("Parse filename error.")
				continue
			}
			if cur > lastest {
				lastest = cur
			}
		}
	}

	// restore
	if lastest > 0 {
		var snapPath string
		incPaths := make(map[int]string)
		incPathsKeys := make([]int, 0)
		todel := make([]string, 0)
		for _, f := range files {
			if f.IsDir() {
				continue
			}

			ss := strings.Split(f.Name(), ".")
			if len(ss) != 2 {
				continue
			}

			ts, err := strconv.Atoi(ss[0])
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err":      err,
					"filename": f.Name(),
				}).Error("Parse filename error.")
				continue
			}

			if ts < lastest {
				todel = append(todel, filepath.Join(pc.conf.SaveDir, f.Name()))
				continue
			}

			switch ss[1] {
			case "snap":
				snapPath = filepath.Join(pc.conf.SaveDir, f.Name())
			case "inc":
				if _, ok := incPaths[ts]; ok {
					logging.VLog().WithFields(logrus.Fields{
						"timestamp": ts,
					}).Error("Duplicated inc file found.")
				} else {
					incPathsKeys = append(incPathsKeys, ts)
					incPaths[ts] = filepath.Join(pc.conf.SaveDir, f.Name())
				}
			case "done":
			}
		}

		// delete old files
		pc.removeFiles(todel)

		// restore cache from files
		pc.restoreSnap(snapPath)
		if len(incPathsKeys) > 0 {
			sort.Ints(incPathsKeys)
			sortedPaths := make([]string, 0)
			for _, k := range incPathsKeys {
				sortedPaths = append(sortedPaths, incPaths[k])
			}
			pc.restoreReplay(sortedPaths)
		}
	}

}

func (pc *PersistableCache) replayToShadow(r *replay) bool {
	switch r.op {
	case Add:
		if _, ok := pc.shadowCache[r.k]; !ok {
			pc.shadowCache[r.k] = r.v
			return true
		}
	case Delete:
		if _, ok := pc.shadowCache[r.k]; ok {
			delete(pc.shadowCache, r.k)
			return true
		}
	}
	// invalid operation
	return false
}

func (pc *PersistableCache) recordReplay(r *replay) {
	var l *SerializableLog // TODO: check nil
	switch r.op {
	case Add:
		kv, err := pc.Encode(r.k, r.v)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Encode replay error.")
		} else {
			l = &SerializableLog{
				Add,
				kv,
			}
		}
	case Delete:
		kv, err := pc.Encode(r.k, nil)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Encode replay error.")
		} else {
			l = &SerializableLog{
				Delete,
				kv,
			}
		}
	default:
		logging.VLog().WithFields(logrus.Fields{
			"type": r.op,
		}).Error("Unknown replay type.")
	}

	if l != nil {
		err := pc.incEncoder.Encode(l)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
				"key": l.Operand.K,
			}).Error("Failed to record replay.")
		}
	}
}

func (pc *PersistableCache) restoreSnap(path string) {
	if path == "" {
		return
	}
	f, err := os.Open(path)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":  err,
			"file": path,
		}).Fatal("Open file error.") // TODO: Fatal?
	}

	decoder := gob.NewDecoder(f)
	for {
		e := new(ExportableEntry)
		err = decoder.Decode(e)
		if err == io.EOF {
			break
		}
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"file": path,
			}).Error("Gob decode error.")
			continue
		}
		if reflect.ValueOf(e.K).IsNil() {
			logging.VLog().WithFields(logrus.Fields{
				"file": path,
			}).Error("ExportableEntry.K is nil.")
			continue
		}

		k, v, err := pc.Decode(e)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"file": path,
			}).Error("Cache entry decode error.")
			continue
		}
		err = pc.Set(k, v)
		// err = pc.originCache.Set(k, v)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Restore error.")
		}
	}
	err = f.Close()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":  err,
			"file": path,
		}).Error("Close file error.")
	}
}

func (pc *PersistableCache) restoreReplay(paths []string) {
	if len(paths) == 0 {
		return
	}

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"file": path,
			}).Fatal("Open file error.") // TODO: Fatal?
		}

		decoder := gob.NewDecoder(f)
		for {
			r := new(SerializableLog)
			err = decoder.Decode(r)
			if err == io.EOF {
				break
			}
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err":  err,
					"file": path,
				}).Error("Gob decode error.")
				continue
			}

			k, v, err := pc.Decode(r.Operand)
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err":  err,
					"file": path,
				}).Error("Replay log decode error.")
				continue
			}

			switch r.Op {
			case Add:
				err = pc.Set(k, v)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err":  err,
						"file": path,
					}).Error("Replay set error.")
				}
			case Delete:
				_, err = pc.Take(k)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err":  err,
						"file": path,
					}).Error("Replay delete error.")
				}
			}
		}

		err = f.Close()
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"file": path,
			}).Error("Close file error.")
		}
	}
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

	if last != "" {
		// delete last files
		pc.removeFiles([]string{
			path.Join(pc.conf.SaveDir, last+".snap"),
			path.Join(pc.conf.SaveDir, last+".done"),
			path.Join(pc.conf.SaveDir, last+".inc"),
		})
	}
}

func (pc *PersistableCache) changeReplayLogger(now string) {
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

func (pc *PersistableCache) removeFiles(paths []string) {
	// TODO: test
	for _, path := range paths {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"file": path,
			}).Error("Stat file error.")
			continue
		}

		err = os.Remove(path)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"file": path,
			}).Error("Delete file error.")
		} else {
			logging.VLog().WithFields(logrus.Fields{
				"file": path,
			}).Info("Delete file.")
		}
	}
}
