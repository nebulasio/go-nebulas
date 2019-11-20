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
	"fmt"
	"os"
	"os/signal"
	"strings"

	"io"

	"bytes"
	"encoding/json"

	"github.com/nebulasio/go-nebulas/cmd/console/library"
	nebletpb "github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/peterh/liner"
)

var (
	defaultPrompt = "> "

	exitCmd = "exit"

	bignumberJS = library.MustAsset("bignumber.js")
	nebJS       = library.MustAsset("neb-light.js")
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() *nebletpb.Config
}

// Console console handler
type Console struct {

	// terminal input prompter
	prompter UserPrompter

	// Channel to send the next prompt on and receive the input
	promptCh chan string

	// input history
	history []string

	// js bridge with go func
	jsBridge *jsBridge

	// js runtime environment
	jsvm *JSVM

	// output writer
	writer io.Writer
}

// Config neb console config
type Config struct {
	Prompter   UserPrompter
	PrompterCh chan string
	Writer     io.Writer
	Neb        Neblet
}

// New a console by Config, neb.config params is need
func New(conf Config) *Console {
	c := new(Console)

	if conf.Prompter != nil {
		c.prompter = conf.Prompter
	}

	if conf.PrompterCh != nil {
		c.promptCh = conf.PrompterCh
	}

	if conf.Writer != nil {
		c.writer = conf.Writer
	}

	if conf.Neb != nil {
		c.jsBridge = newBirdge(conf.Neb.Config(), c.prompter, c.writer)
	} else {
		// hack for test with local environment
		c.jsBridge = newBirdge(nil, c.prompter, c.writer)
	}

	c.jsvm = newJSVM()
	if err := c.loadLibraryScripts(); err != nil {
		fmt.Fprintln(c.writer, err)
	}

	if err := c.methodSwizzling(); err != nil {
		fmt.Fprintln(c.writer, err)
	}
	return c
}

func (c *Console) loadLibraryScripts() error {
	if err := c.jsvm.Compile("bignumber.js", bignumberJS); err != nil {
		return fmt.Errorf("bignumber.js: %v", err)
	}
	if err := c.jsvm.Compile("neb-light.js", nebJS); err != nil {
		return fmt.Errorf("neb.js: %v", err)
	}
	return nil
}

// Individual methods use go implementation
func (c *Console) methodSwizzling() error {

	// replace js console log & error with go impl
	jsconsole, _ := c.jsvm.Get("console")
	jsconsole.Object().Set("log", c.jsBridge.output)
	jsconsole.Object().Set("error", c.jsBridge.output)

	// replace js xmlhttprequest to go implement
	c.jsvm.Set("bridge", struct{}{})
	bridgeObj, _ := c.jsvm.Get("bridge")
	bridgeObj.Object().Set("request", c.jsBridge.request)
	bridgeObj.Object().Set("asyncRequest", c.jsBridge.request)

	if _, err := c.jsvm.Run("var Neb = require('neb');"); err != nil {
		return fmt.Errorf("neb require: %v", err)
	}
	if _, err := c.jsvm.Run("var neb = new Neb(bridge);"); err != nil {
		return fmt.Errorf("neb create: %v", err)
	}
	jsAlias := "var api = neb.api; var admin = neb.admin; "
	if _, err := c.jsvm.Run(jsAlias); err != nil {
		return fmt.Errorf("namespace: %v", err)
	}

	if c.prompter != nil {
		admin, err := c.jsvm.Get("admin")
		if err != nil {
			return err
		}
		if obj := admin.Object(); obj != nil {
			bridgeRequest := `bridge._sendRequest = function (method, api, params, callback) {
				var action = "/admin" + api;
				return this.request(method, action, params);
			};`
			if _, err = c.jsvm.Run(bridgeRequest); err != nil {
				return fmt.Errorf("bridge._sendRequest: %v", err)
			}

			if _, err = c.jsvm.Run(`bridge.newAccount = admin.newAccount;`); err != nil {
				return fmt.Errorf("admin.newAccount: %v", err)
			}
			if _, err = c.jsvm.Run(`bridge.unlockAccount = admin.unlockAccount;`); err != nil {
				return fmt.Errorf("admin.unlockAccount: %v", err)
			}
			if _, err = c.jsvm.Run(`bridge.sendTransactionWithPassphrase = admin.sendTransactionWithPassphrase;`); err != nil {
				return fmt.Errorf("admin.sendTransactionWithPassphrase: %v", err)
			}
			if _, err = c.jsvm.Run(`bridge.signTransactionWithPassphrase = admin.signTransactionWithPassphrase;`); err != nil {
				return fmt.Errorf("admin.signTransactionWithPassphrase: %v", err)
			}
			obj.Set("setHost", c.jsBridge.setHost)
			obj.Set("newAccount", c.jsBridge.newAccount)
			obj.Set("unlockAccount", c.jsBridge.unlockAccount)
			obj.Set("sendTransactionWithPassphrase", c.jsBridge.sendTransactionWithPassphrase)
			obj.Set("signTransactionWithPassphrase", c.jsBridge.signTransactionWithPassphrase)
		}
	}
	return nil
}

// AutoComplete console auto complete input
func (c *Console) AutoComplete(line string, pos int) (string, []string, string) {
	// No completions can be provided for empty inputs
	if len(line) == 0 || pos == 0 {
		return "", nil, ""
	}
	start := pos - 1
	for ; start > 0; start-- {
		// Skip all methods and namespaces
		if line[start] == '.' || (line[start] >= 'a' && line[start] <= 'z') || (line[start] >= 'A' && line[start] <= 'Z') {
			continue
		}
		start++
		break
	}
	if start == pos {
		return "", nil, ""
	}
	return line[:start], c.jsvm.CompleteKeywords(line[start:pos]), line[pos:]
}

// Setup setup console
func (c *Console) Setup() {
	if c.prompter != nil {
		c.prompter.SetWordCompleter(c.AutoComplete)
	}
	fmt.Fprint(c.writer, "Welcome to the Neb JavaScript console!\n")
}

// Interactive starts an interactive user session.
func (c *Console) Interactive() {
	// Start a goroutine to listen for promt requests and send back inputs
	go func() {
		for {
			// Read the next user input
			line, err := c.prompter.Prompt(<-c.promptCh)
			if err != nil {
				if err == liner.ErrPromptAborted { // ctrl-C
					c.promptCh <- exitCmd
					continue
				}
				close(c.promptCh)
				return
			}
			c.promptCh <- line
		}
	}()
	// Monitor Ctrl-C
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, os.Interrupt, os.Kill)

	// Start sending prompts to the user and reading back inputs
	for {
		// Send the next prompt, triggering an input read and process the result
		c.promptCh <- defaultPrompt
		select {
		case <-abort:
			fmt.Fprint(c.writer, "exiting...")
			return
		case line, ok := <-c.promptCh:
			// User exit
			if !ok || strings.ToLower(line) == exitCmd {
				return
			}
			if len(strings.TrimSpace(line)) == 0 {
				continue
			}

			if command := strings.TrimSpace(line); len(c.history) == 0 || command != c.history[len(c.history)-1] {
				c.history = append(c.history, command)
				if c.prompter != nil {
					c.prompter.AppendHistory(command)
				}
			}
			c.Evaluate(line)
		}
	}
}

// Evaluate executes code and pretty prints the result
func (c *Console) Evaluate(code string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(c.writer, "[native] error: %v\n", r)
		}
	}()
	v, err := c.jsvm.Run(code)
	if err != nil {
		fmt.Fprintln(c.writer, err)
		return err
	}
	if v.IsObject() {
		result, err := c.jsvm.JSONString(v)
		if err != nil {
			fmt.Fprintln(c.writer, err)
			return err
		}
		var buf bytes.Buffer
		err = json.Indent(&buf, []byte(result), "", "    ")
		if err != nil {
			fmt.Fprintln(c.writer, err)
			return err
		}
		fmt.Fprintln(c.writer, buf.String())
	} else if v.IsString() {
		fmt.Fprintln(c.writer, v.String())
	}
	return nil
}

// Stop stop js console
func (c *Console) Stop() error {
	return nil
}
