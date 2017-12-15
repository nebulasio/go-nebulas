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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// LogrusFileHook file hook for stdout
type LogrusFileHook struct {
	file      *os.File
	flag      int
	chmod     os.FileMode
	formatter *logrus.TextFormatter
}

// NewLogrusFileHook create logrus file hook
func NewLogrusFileHook(dir string) (*LogrusFileHook, error) {
	if !filepath.IsAbs(dir) {
		dir, _ = filepath.Abs(dir)
	}
	flag := os.O_CREATE | os.O_APPEND | os.O_RDWR
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	file := dir + "/neb-" + time.Now().Format("2006-01-02") + ".log"
	plainFormatter := &logrus.TextFormatter{DisableColors: true}
	logFile, err := os.OpenFile(file, flag, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to write file on filehook %v", err)
		return nil, err
	}

	return &LogrusFileHook{logFile, flag, 0666, plainFormatter}, err
}

// Fire event
func (hook *LogrusFileHook) Fire(entry *logrus.Entry) error {

	plainformat, err := hook.formatter.Format(entry)
	line := string(plainformat)
	_, err = hook.file.WriteString(line)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to write file on filehook(entry.String)%v", err)
		return err
	}

	return nil
}

// Levels log levels
func (hook *LogrusFileHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
