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
	"os"

	"github.com/sirupsen/logrus"
)

// const
const (
	PanicLevel = "panic"
	FatalLevel = "fatal"
	ErrorLevel = "error"
	WarnLevel  = "warn"
	InfoLevel  = "info"
	DebugLevel = "debug"
)

type emptyWriter struct{}

func (ew emptyWriter) Write(p []byte) (int, error) {
	return 0, nil
}

var clog *logrus.Logger
var vlog *logrus.Logger

// CLog return console logger
func CLog() *logrus.Logger {
	if clog == nil {
		Init("/tmp", "info", 0)
	}
	return clog
}

// VLog return verbose logger
func VLog() *logrus.Logger {
	if vlog == nil {
		Init("/tmp", "info", 0)
	}
	return vlog
}

func convertLevel(level string) logrus.Level {
	switch level {
	case PanicLevel:
		return logrus.PanicLevel
	case FatalLevel:
		return logrus.FatalLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	case WarnLevel:
		return logrus.WarnLevel
	case InfoLevel:
		return logrus.InfoLevel
	case DebugLevel:
		return logrus.DebugLevel
	default:
		return logrus.InfoLevel
	}
}

// Init loggers
func Init(path string, level string, age uint32) {
	fileHooker := NewFileRotateHooker(path, age)

	clog = logrus.New()
	LoadFunctionHooker(clog)
	clog.Hooks.Add(fileHooker)
	clog.Out = os.Stdout
	clog.Formatter = &logrus.TextFormatter{FullTimestamp: true}
	clog.Level = convertLevel("debug")

	vlog = logrus.New()
	LoadFunctionHooker(vlog)
	vlog.Hooks.Add(fileHooker)
	vlog.Out = &emptyWriter{}
	vlog.Formatter = &logrus.TextFormatter{FullTimestamp: true}
	vlog.Level = convertLevel(level)

	VLog().WithFields(logrus.Fields{
		"path":  path,
		"level": level,
	}).Info("Logger Configuration.")
}
