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

// const
const (
	SnapFileSuffix     = ".snap"
	SnapDoneFileSuffix = ".done"
	ReplayFileSuffix   = ".rep"
)

type replay struct {
	op Operator
	k  interface{}
	v  interface{}
}

// PersistableCache ..
type PersistableCache struct {
	mirrorCache        map[interface{}]interface{}
	originCache        Cache
	conf               *PersistableCacheConfig
	replayFile         *os.File
	replayEncoder      *gob.Encoder
	replayChan         chan *replay
	persistenceStarted bool
	ticker             *time.Ticker
}

// NewPersistableCache ..
func NewPersistableCache(cache Cache, conf *PersistableCacheConfig) Cache {
	pc := &PersistableCache{
		mirrorCache:        make(map[interface{}]interface{}),
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

// EncodeEntry ..
func (pc *PersistableCache) EncodeEntry(k, v interface{}) (*ExportableEntry, error) {
	return pc.originCache.EncodeEntry(k, v)
}

// DecodeEntry ..
func (pc *PersistableCache) DecodeEntry(kv *ExportableEntry) (interface{}, interface{}, error) {
	return pc.originCache.DecodeEntry(kv)
}

// StartPersistence ..
func (pc *PersistableCache) StartPersistence() {
	if pc.persistenceStarted {
		return
	}
	// dump immediately
	now := strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond)))
	pc.nextReplayLogger(now)
	pc.dumpMirror(now, 0)
	pc.ticker = time.NewTicker(time.Millisecond * time.Duration(pc.conf.IntervalMs))

	pc.persistenceStarted = true
}

func (pc *PersistableCache) init() {

	pc.restore()

	go func() {
		last := 0
		for r := range pc.replayChan {

			b := pc.replayToMirror(r)

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
				pc.nextReplayLogger(now)
				pc.dumpMirror(now, last)
				end := int(time.Now().UnixNano() / int64(time.Millisecond))

				logging.VLog().WithFields(logrus.Fields{
					"cost":      end - start,
					"dir":       pc.conf.SaveDir,
					"timestamp": now,
				}).Info("Dump snapshot done.")

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
			}).Fatal("Create cache dir error.")
		}
		return
	}
	if !fi.IsDir() {
		logging.VLog().WithFields(logrus.Fields{
			"file": pc.conf.SaveDir,
		}).Fatal("File already exists with the same name.")
	}

	// list files
	files, err := ioutil.ReadDir(pc.conf.SaveDir)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"dir": pc.conf.SaveDir,
		}).Fatal("List files error.")
	}

	// pick max timestamp
	lastest := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasSuffix(f.Name(), SnapDoneFileSuffix) {
			cur, err := strconv.Atoi(strings.Split(f.Name(), SnapDoneFileSuffix)[0])
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
		repPaths := make(map[int]string)
		repPathsKeys := make([]int, 0)
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
			case "rep":
				repPathsKeys = append(repPathsKeys, ts)
				repPaths[ts] = filepath.Join(pc.conf.SaveDir, f.Name())
			default:
			}
		}

		// delete old files
		pc.removeFiles(todel)

		// restore cache from files
		pc.restoreSnap(snapPath)
		if len(repPathsKeys) > 0 {
			sort.Ints(repPathsKeys)
			sortedPaths := make([]string, 0)
			for _, k := range repPathsKeys {
				sortedPaths = append(sortedPaths, repPaths[k])
			}
			pc.restoreReplay(sortedPaths)
		}
	}

}

func (pc *PersistableCache) replayToMirror(r *replay) bool {
	switch r.op {
	case Add:
		if _, ok := pc.mirrorCache[r.k]; !ok {
			pc.mirrorCache[r.k] = r.v
			return true
		}
	case Delete:
		if _, ok := pc.mirrorCache[r.k]; ok {
			delete(pc.mirrorCache, r.k)
			return true
		}
	}
	// invalid operation
	return false
}

func (pc *PersistableCache) recordReplay(r *replay) {
	var l *ReplayLog // TODO: check nil
	switch r.op {
	case Add:
		kv, err := pc.EncodeEntry(r.k, r.v)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Encode add replay error.")
		} else {
			l = &ReplayLog{
				Add,
				kv,
			}
		}
	case Delete:
		kv, err := pc.EncodeEntry(r.k, nil)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Encode delete replay error.")
		} else {
			l = &ReplayLog{
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
		err := pc.replayEncoder.Encode(l)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"type": l.Op,
				"key":  l.Operand.K,
			}).Error("Failed to record replay.")
		}
	}
}

func (pc *PersistableCache) restoreSnap(path string) {

	f, err := os.Open(path)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":  err,
			"file": path,
		}).Fatal("Open snap file error.") // TODO: Fatal?
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

		k, v, err := pc.DecodeEntry(e)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"file": path,
			}).Error("Decode cache entry error.")
			continue
		}
		err = pc.Set(k, v)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"file": path,
			}).Error("Restore cache entry error.")
		}
	}
	err = f.Close()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":  err,
			"file": path,
		}).Error("Close snap file error.")
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
			}).Fatal("Open replay file error.") // TODO: Fatal?
		}

		decoder := gob.NewDecoder(f)
		for {
			r := new(ReplayLog)
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

			k, v, err := pc.DecodeEntry(r.Operand)
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err":  err,
					"file": path,
				}).Error("Decode replay entry error.")
				continue
			}

			switch r.Op {
			case Add:
				err = pc.Set(k, v)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err":  err,
						"file": path,
					}).Error("Restore add error.")
				}
			case Delete:
				_, err = pc.Take(k)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err":  err,
						"file": path,
					}).Error("Restore delete error.")
				}
			default:
				logging.VLog().WithFields(logrus.Fields{
					"type": r.Op,
					"file": path,
				}).Error("Unknown replay log type.")
			}
		}

		err = f.Close()
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"file": path,
			}).Error("Close replay file error.")
		}
	}
}

func (pc *PersistableCache) dumpMirror(now string, last int) {

	p := path.Join(pc.conf.SaveDir, now+SnapFileSuffix)
	f, err := os.Create(p)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":       err,
			"timestamp": now,
		}).Fatal("Create snapshot file error.") // TODO: Fatal?
	}
	snapEncoder := gob.NewEncoder(f)

	for k, v := range pc.mirrorCache {
		kv, err := pc.EncodeEntry(k, v)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Serialize ExportableEntry error.")
			continue
		}
		err = snapEncoder.Encode(kv)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
				"key": k,
			}).Error("Encode ExportableEntry error.")
		}
	}

	err = f.Close()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":  err,
			"file": p,
		}).Error("Close snapshot file error.")
	}

	// create done file
	doneF, err := os.Create(path.Join(pc.conf.SaveDir, now+SnapDoneFileSuffix))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":       err,
			"timestamp": now,
		}).Fatal("Create done file error.") // TODO: Fatal?
	} else {
		defer doneF.Close()
	}

	if last > 0 {
		// delete last files
		sl := strconv.Itoa(last)
		pc.removeFiles([]string{
			path.Join(pc.conf.SaveDir, sl+SnapFileSuffix),
			path.Join(pc.conf.SaveDir, sl+SnapDoneFileSuffix),
			path.Join(pc.conf.SaveDir, sl+ReplayFileSuffix),
		})
	}
}

func (pc *PersistableCache) nextReplayLogger(now string) {
	if pc.replayFile != nil {
		err := pc.replayFile.Close()
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"file": pc.replayFile.Name(),
			}).Error("Close last replay file error.")
		}
	}

	p := filepath.Join(pc.conf.SaveDir, now+ReplayFileSuffix)
	replayFile, err := os.Create(p)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":      err,
			"filepath": p,
		}).Fatal("Create replay file error.") // TODO: Fatal?
	}

	pc.replayFile = replayFile
	pc.replayEncoder = gob.NewEncoder(replayFile)

	logging.VLog().WithFields(logrus.Fields{
		"file": pc.replayFile.Name(),
	}).Info("Create new replay file done.")
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
			}).Info("Delete file done.")
		}
	}
}
