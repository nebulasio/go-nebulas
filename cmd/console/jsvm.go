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

package console

import (
	"github.com/robertkrimen/otto"
)

// JSVM javascript runtime environment
type JSVM struct {

	// the representation of the JavaScript runtime
	vm *otto.Otto
}

func newJSVM() *JSVM {
	vm := &JSVM{}
	vm.vm = otto.New()
	return vm
}

// Run will run the given source (parsing it first if necessary), returning the resulting value and error (if any)
func (v *JSVM) Run(src string) (otto.Value, error) {
	return v.vm.Run(src)
}

// Get returns the value of a variable in the JS environment.
func (v *JSVM) Get(name string) (otto.Value, error) {
	return v.vm.Get(name)
}

// Set assigns value v to a variable in the JS environment.
func (v *JSVM) Set(name string, value interface{}) error {
	return v.vm.Set(name, value)
}

// Compile compiles and then runs JS code.
func (v *JSVM) Compile(filename string, src interface{}) error {
	script, err := v.vm.Compile(filename, src)
	if err != nil {
		return err
	}
	_, err = v.vm.Run(script)
	return err
}
