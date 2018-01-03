// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package logging

import (
	"path/filepath"
	"runtime"
	"strings"

	logrus "github.com/sirupsen/logrus"
)

type functionHooker struct {
	innerLogger *logrus.Logger
	file        string
}

func (h *functionHooker) Fire(entry *logrus.Entry) error {
	pc := make([]uintptr, 10)
	runtime.Callers(6, pc)
	for i := 0; i < 10; i++ {
		if pc[i] == 0 {
			break
		}
		f := runtime.FuncForPC(pc[i])
		file, line := f.FileLine(pc[i])
		if strings.Contains(file, "sirupsen") {
			continue
		}
		fname := f.Name()
		if strings.Contains(fname, "/") {
			index := strings.LastIndex(fname, "/")
			entry.Data["func"] = fname[index+1:]
			// entry.Data["package"] = fname[0:index]
		} else {
			entry.Data["func"] = fname
		}
		entry.Data["line"] = line
		entry.Data["file"] = filepath.Base(file)
		break
	}
	return nil
}

func (h *functionHooker) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

// LoadFunctionHooker loads a function hooker to the logger
func LoadFunctionHooker(logger *logrus.Logger) {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
	}
	inst := &functionHooker{
		innerLogger: logger,
		file:        file,
	}
	logger.Hooks.Add(inst)
}
