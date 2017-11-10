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
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/robertkrimen/otto"
)

type jsBridge struct {

	// terminal input prompter
	prompter *terminalPrompter

	writer io.Writer
}

// newBirdge create a new jsbridge with given prompter and writer
func newBirdge(config nebletpb.Config, prompter *terminalPrompter, writer io.Writer) *jsBridge {
	bridge := &jsBridge{prompter: prompter, writer: writer}
	return bridge
}

// output handle the error & log in js runtime
func (b *jsBridge) output(call otto.FunctionCall) {
	output := []string{}
	for _, argument := range call.ArgumentList {
		output = append(output, fmt.Sprintf("%v", argument))
	}
	fmt.Fprintln(b.writer, strings.Join(output, " "))
}

// request handle http request
func (b *jsBridge) request(call otto.FunctionCall) otto.Value {
	return otto.NullValue()
}

// newAccount handle the account generate with passphrase input
func (b *jsBridge) newAccount(call otto.FunctionCall) otto.Value {
	var (
		password string
		err      error
	)
	switch {
	// No password was specified, prompt the user for it
	case len(call.ArgumentList) == 0:
		if password, err = b.prompter.PromptPassphrase("Passphrase: "); err != nil {
			fmt.Fprint(b.writer, err)
			return otto.NullValue()
		}
		var confirm string
		if confirm, err = b.prompter.PromptPassphrase("Repeat passphrase: "); err != nil {
			fmt.Fprint(b.writer, err)
			return otto.NullValue()
		}
		if password != confirm {
			fmt.Fprint(b.writer, errors.New("passphrase don't match"))
			return otto.NullValue()
		}
	case len(call.ArgumentList) == 1 && call.Argument(0).IsString():
		password, _ = call.Argument(0).ToString()
	default:
		fmt.Fprint(b.writer, errors.New("unexpected argument count"))
		return otto.NullValue()
	}
	ret, err := call.Otto.Call("neb.newAccount", nil, password)
	if err != nil {
		fmt.Fprint(b.writer, err)
		return otto.NullValue()
	}
	return ret
}

// signTransaction handle the account unlock with passphrase input
func (b *jsBridge) unlockAccount(call otto.FunctionCall) otto.Value {
	if !call.Argument(0).IsString() {
		fmt.Fprint(b.writer, errors.New("address arg must be string"))
		return otto.NullValue()
	}
	address := call.Argument(0)

	var passphrase otto.Value

	if call.Argument(1).IsUndefined() || call.Argument(1).IsNull() {
		fmt.Fprintf(b.writer, "Unlock account %s\n", address)
		var (
			input string
			err   error
		)
		if input, err = b.prompter.PromptPassphrase("Passphrase: "); err != nil {
			fmt.Fprint(b.writer, err)
			return otto.NullValue()
		}
		passphrase, _ = otto.ToValue(input)
	} else {
		if !call.Argument(1).IsString() {
			fmt.Fprint(b.writer, errors.New("password must be a string"))
			return otto.NullValue()
		}
		passphrase = call.Argument(1)
	}

	// Send the request to the backend and return
	val, err := call.Otto.Call("jeth.unlockAccount", nil, address, passphrase)
	if err != nil {
		fmt.Fprint(b.writer, err)
		return otto.NullValue()
	}
	return val
}

//// signTransaction handle the transaction sign with passphrase input
//func (b *jsBridge)signTransaction(call otto.FunctionCall) otto.Value {
//	return nil
//}

//// sendTransaction handle the transaction send with passphrase input
//func (b *jsBridge)sendTransactionWithPassphrase(call otto.FunctionCall) otto.Value {
//	return nil
//}
