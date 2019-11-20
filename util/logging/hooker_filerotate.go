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
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// NewFileRotateHooker enable log file output
func NewFileRotateHooker(path string, age uint32) logrus.Hook {
	if len(path) == 0 {
		panic("Failed to parse logger folder:" + path + ".")
	}
	if !filepath.IsAbs(path) {
		path, _ = filepath.Abs(path)
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		panic("Failed to create logger folder:" + path + ". err:" + err.Error())
	}
	filePath := path + "/neb-%Y%m%d-%H.log"
	linkPath := path + "/neb.log"
	writer, err := rotatelogs.New(
		filePath,
		rotatelogs.WithLinkName(linkPath),
		rotatelogs.WithRotationTime(time.Duration(3600)*time.Second),
		rotatelogs.WithMaxAge(time.Duration(age)*time.Second),
	)
	if err != nil {
		panic("Failed to create rotate logs. err:" + err.Error())
	}

	hook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
	}, nil)
	return hook
}
