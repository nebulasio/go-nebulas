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

package util

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	// ErrFileExists file exists
	ErrFileExists = errors.New("file exists")
)

// CreateDirIfNotExist create dir
func CreateDirIfNotExist(dir string) error {
	if exist, err := FileExists(dir); !exist || err != nil {
		if err != nil {
			return err
		}
		err = os.MkdirAll(dir, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// FileExists check file exists
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// FileWrite write file to path
func FileWrite(file string, content []byte, overwrite bool) error {
	// Create the keystore directory with appropriate permissions
	if err := CreateDirIfNotExist(filepath.Dir(file)); err != nil {
		return err
	}
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()

	if overwrite {
		if exist, _ := FileExists(file); exist {
			if err := os.Remove(file); err != nil {
				os.Remove(f.Name())
				return err
			}
		}
	}

	return os.Rename(f.Name(), file)
}
