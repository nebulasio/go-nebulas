package console

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/peterh/liner"
)

type tester struct {
	console *Console
	input   *hookedPrompter
	output  *bytes.Buffer
}

type hookedPrompter struct {
	scheduler chan string
}

// Prompt shows the prompt and requests text input
// returning the input.
func (p *hookedPrompter) Prompt(prompt string) (string, error) {
	// Send the prompt to the tester
	select {
	case p.scheduler <- prompt:
	case <-time.After(time.Second):
		return "", errors.New("prompt timeout")
	}
	// Retrieve the response and feed to the console
	select {
	case input := <-p.scheduler:
		return input, nil
	case <-time.After(time.Second):
		return "", errors.New("input timeout")
	}
}

// PromptPassphrase shows the prompt and request passphrase text input, the passphrase
// not show, returns the passphrase
func (p *hookedPrompter) PromptPassphrase(prompt string) (passwd string, err error) {
	return "", nil
}

// PromptConfirm shows the prompt to the user and requests a boolean
// choice to be made, returning that choice.
func (p *hookedPrompter) PromptConfirm(prompt string) (bool, error) {
	input, err := p.Prompt(prompt + " [y/N] ")
	if len(input) > 0 && strings.ToUpper(input[:1]) == "Y" {
		return true, nil
	}
	return false, err
}

// SetHistory sets the history that the prompter will allow
// the user to scroll back to.
func (p *hookedPrompter) SetHistory(history []string) {
	//fmt.Println("to be implemented sethistory")
}

// AppendHistory appends an entry to the scrollback history.
func (p *hookedPrompter) AppendHistory(command string) {
	//fmt.Println("to be implemented AppendHistory")
}

// SetWordCompleter sets the completion function that the prompter will call to
// fetch completion candidates when the user presses tab.
func (p *hookedPrompter) SetWordCompleter(completer liner.WordCompleter) {
	//fmt.Println("to be implemented SetWordCompleter")
}

func (env *tester) Close(t *testing.T) {
}

func newTester(t *testing.T) *tester {

	hookedInput := &hookedPrompter{scheduler: make(chan string)}
	hookedOutput := new(bytes.Buffer)

	testConsole := New(Config{
		Prompter:   hookedInput,
		PrompterCh: make(chan string),
		Writer:     hookedOutput,
		Neb:        nil,
	})

	return &tester{
		console: testConsole,
		input:   hookedInput,
		output:  hookedOutput,
	}
}

// Tests that the console can be used in interactive mode
func TestConsoleInteractive(t *testing.T) {
	tester := newTester(t)
	defer tester.Close(t)
	go tester.console.Interactive()

	// Wait for a promt and send a statement back
	select {
	case <-tester.input.scheduler:
	case <-time.After(time.Second):
		t.Fatalf("initial prompt timeout")
	}

	select {
	case tester.input.scheduler <- "obj = {expect: 2+2}":
	case <-time.After(time.Second):
		t.Fatalf("input feedback timeout")
	}

	// Wait for the second promt and ensure first statement was evaluated
	select {
	case <-tester.input.scheduler:
	case <-time.After(time.Second):
		t.Fatalf("thirt prompt timeout")
	}

	output := tester.output.String()
	if !strings.Contains(output, "4") {
		t.Fatalf("statement evaluation failed: have %s, want %s", output, "4")
	}
}

// Test pre cmd can be run
func TestRun(t *testing.T) {
	tester := newTester(t)
	defer tester.Close(t)
	tester.console.jsvm.Run(`var preloaded_var = "some_preloaded_var"`)
	tester.console.Evaluate("preloaded_var")
	output := tester.output.String()
	if !strings.Contains(output, "some_preloaded_var") {
		t.Fatalf("preloaded variable missing: have %s, want %s", output, "some-preloaded-string")
	}
}

// Tests that JavaScript statement evaluation works as intended.
func TestEvaluate(t *testing.T) {
	tester := newTester(t)
	defer tester.Close(t)

	tester.console.Evaluate("'abc'+'bcd'")
	if output := tester.output.String(); !strings.Contains(output, "abcbcd") {
		t.Fatalf("statement evaluation failed: have %s, want %s", output, "abcbcd")
	}
}

// Tests that the JavaScript objects returned by statement executions are properly
// pretty printed instead of just displaying "[object]".

func TestPrettyPrint(t *testing.T) {
	tester := newTester(t)
	defer tester.Close(t)

	tester.console.Evaluate("obj = {int: 1, string: 'i am a string', list: [3, 3], obj: {null:null}}")

	// Assemble the actual output we're after and verify
	want := `{
    "int": 1,
    "list": [
        3,
        3
    ],
    "obj": {
        "null": null
    },
    "string": "i am a string"
}
`
	if output := tester.output.String(); output != want {
		t.Fatalf("pretty print mismatch: have %s, want %s", output, want)
	}
}

// Tests that the console can handle valid commands (no arguments)
func TestApiValidInput(t *testing.T) {
	tester := newTester(t)
	defer tester.Close(t)
	go tester.console.Interactive()

	testCases := []struct {
		inputCommand               string
		expectOutputWhenNebRunning string
		// when local neb is not started only get expectOutputWhenNebNotRunning
		expectOutputWhenNebNotRunning string
	}{
		/* {`admin.accounts()`, "\"addresses\"", "connection refuse"}, */ // TODO: @cheng recover
		/* {`admin.nodeInfo()`, "\"bucket_size\"", "connection refuse"}, */
		/* {`api.blockDump()`, "\"data\"", "connection refuse"}, */
		{`api.gasPrice()`, "\"gas_price\"", "connection refuse"},
		{`api.getNebState()`, "\"chain_id\"", "connection refuse"},
	}

	// Wait for a promt and send a statement back
	select {
	case <-tester.input.scheduler:
	case <-time.After(time.Second):
		t.Fatalf("initial prompt timeout")
	}

	for _, tt := range testCases {
		fmt.Println("testing " + tt.inputCommand)

		select {
		case tester.input.scheduler <- tt.inputCommand:
		case <-time.After(time.Second):
			t.Fatalf("input feedback timeout")
		}
		// Wait for the second promt and ensure first statement was evaluated
		select {
		case <-tester.input.scheduler:
		case <-time.After(time.Second * 2):
			t.Fatalf("secondary prompt timeout")
		}

		output := tester.output.String()

		// Reset resets the buffer to make it has no content
		tester.output.Reset()

		if strings.Contains(output, tt.expectOutputWhenNebRunning) {
			continue
		} else if strings.Contains(output, tt.expectOutputWhenNebNotRunning) {
			fmt.Println("testing console without running neb")
		} else {
			t.Fatalf("statement evaluation failed: have %s, want %s", output, tt.expectOutputWhenNebRunning+" or "+tt.expectOutputWhenNebNotRunning)
		}
	}
}

func TestSpecialInput(t *testing.T) {
	tester := newTester(t)
	defer tester.Close(t)
	go tester.console.Interactive()

	testCases := []struct {
		inputCommand               string
		expectOutputWhenNebRunning string
		// when local neb is not started only get expectOutputWhenNebNotRunning
		expectOutputWhenNebNotRunning string
	}{
		{`#(*$)@(*#$)(*$`, `\"Unexpected token ILLEGAL\"`, "Unexpected token ILLEGAL"},
		{``, "", ""},
		{`alert("Hello! I am an alert box!!");`, "alert' is not defined", "alert' is not defined"},
		{`admin.unknownCommand();`, "unknownCommand' is not a function", "unknownCommand' is not a function"},
	}
	// Wait for a promt and send a statement back
	select {
	case <-tester.input.scheduler:
	case <-time.After(time.Second):
		t.Fatalf("initial prompt timeout")
	}

	for _, tt := range testCases {
		fmt.Println("testing " + tt.inputCommand)

		select {
		case tester.input.scheduler <- tt.inputCommand:
		case <-time.After(time.Second):
			t.Fatalf("input feedback timeout")
		}
		// Wait for the second promt and ensure first statement was evaluated
		select {
		case <-tester.input.scheduler:
		case <-time.After(time.Second):
			t.Fatalf("secondary prompt timeout")
		}

		output := tester.output.String()

		// Reset resets the buffer to make it has no content
		tester.output.Reset()

		if strings.Contains(output, tt.expectOutputWhenNebRunning) {
			continue
		} else if strings.Contains(output, tt.expectOutputWhenNebNotRunning) {
			fmt.Println("testing console without running neb")
		} else {
			t.Fatalf("statement evaluation failed: have %s, want %s", output, tt.expectOutputWhenNebRunning+" or "+tt.expectOutputWhenNebNotRunning)
		}
	}
}
