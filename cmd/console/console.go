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

	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/peterh/liner"
)

var (
	defaultPrompt = "> "

	exitCmd = "exit"
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() nebletpb.Config
}

// Console console handler
type Console struct {
	prompter *terminalPrompter

	// Channel to send the next prompt on and receive the input
	promptCh chan string

	// input history
	history []string
}

// New new a console obj,config params is need
func New(neb Neblet) *Console {
	c := new(Console)
	if neb != nil {
		//TODO(larry.wang):add conf for consle
	}
	c.prompter = Stdin
	c.promptCh = make(chan string)
	return c
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
	return line[:start], nil, line[pos:]
}

// Setup setup console
func (c *Console) Setup() {
	if c.prompter != nil {
		c.prompter.SetWordCompleter(c.AutoComplete)
	}
	fmt.Println("Welcome to the Neb JavaScript console!")
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
					c.promptCh <- ""
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
			fmt.Println("exiting...")
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
func (c *Console) Evaluate(statement string) error {
	//TODO(larry.wang):evaluate js cmd
	return nil
}

// Stop stop js console
func (c *Console) Stop() error {
	return nil
}
