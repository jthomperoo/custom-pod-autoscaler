/*
Copyright 2019 The Custom Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// +build unit

package shell_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/shell"
)

type command func(name string, arg ...string) *exec.Cmd

type process func(t *testing.T)

func TestShellProcess(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") != "1" {
		return
	}

	processName := strings.Split(os.Args[3], "=")[1]
	process := processes[processName]

	if process == nil {
		t.Errorf("Process %s not found", processName)
		os.Exit(1)
	}

	process(t)

	// Process should call os.Exit itself, if not exit with error
	os.Exit(1)
}

func fakeExecCommandAndStart(name string, process process) command {
	processes[name] = process
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestShellProcess", "--", fmt.Sprintf("-process=%s", name), command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_TEST_PROCESS=1"}
		cmd.Start()
		return cmd
	}
}

func fakeExecCommand(name string, process process) command {
	processes[name] = process
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestShellProcess", "--", fmt.Sprintf("-process=%s", name), command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_TEST_PROCESS=1"}
		return cmd
	}
}

func stringPointer(value string) *string {
	return &value
}

type test struct {
	description   string
	expectedErr   error
	expected      *string
	commandString string
	value         string
	timeout       int
	command       command
}

var tests []test

var processes map[string]process

func TestMain(m *testing.M) {
	processes = map[string]process{}
	tests = []test{
		{
			"Successful shell command",
			nil,
			stringPointer("test std out"),
			"command",
			"pipe value",
			100,
			fakeExecCommand("success", func(t *testing.T) {
				// Check provided values are correct
				// Due to the fake shell command, the actual command and value piped to it
				// are arguments passed to this command - at argument position 6
				// e.g. echo 'stdin' | command
				commandAndPipe := os.Args[6]

				stdin := strings.TrimSpace(strings.Split(commandAndPipe, "|")[0])
				command := strings.TrimSpace(strings.Split(commandAndPipe, "|")[1])

				// stdin is echoed and piped to the command, so the argument will be surrounded
				// by an echo command
				testPipeValueWithEcho := fmt.Sprintf("echo '%s'", "pipe value")

				// Check command is correct
				if !cmp.Equal(command, "command") {
					fmt.Fprintf(os.Stderr, "stdin mismatch (-want +got):\n%s", cmp.Diff("command", command))
					os.Exit(1)
				}

				// Check piped value in is correct
				if !cmp.Equal(stdin, testPipeValueWithEcho) {
					fmt.Fprintf(os.Stderr, "stdin mismatch (-want +got):\n%s", cmp.Diff(testPipeValueWithEcho, stdin))
					os.Exit(1)
				}

				fmt.Fprint(os.Stdout, "test std out")
				os.Exit(0)
			}),
		},
		{
			"Failed shell command",
			errors.New("exit status 1"),
			stringPointer("shell command failed"),
			"command",
			"pipe value",
			100,
			fakeExecCommand("failed", func(t *testing.T) {
				fmt.Fprint(os.Stderr, "shell command failed")
				os.Exit(1)
			}),
		},
		{
			"Failed shell command timeout",
			errors.New("Command command timed out"),
			nil,
			"command",
			"pipe value",
			1,
			fakeExecCommand("timeout", func(t *testing.T) {
				fmt.Fprint(os.Stdout, "test std out")
				time.Sleep(10 * time.Millisecond)
				os.Exit(0)
			}),
		},
		{
			"Failed shell command fail to start",
			errors.New("exec: already started"),
			nil,
			"command",
			"pipe value",
			100,
			fakeExecCommandAndStart("fail to start", func(t *testing.T) {
				fmt.Fprint(os.Stdout, "test std out")
				os.Exit(0)
			}),
		},
	}
	code := m.Run()
	os.Exit(code)
}

func TestExecWithValuePipe(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			s := &shell.ExecuteWithPipe{test.command}
			stdout, err := s.ExecuteWithPipe(test.commandString, test.value, test.timeout)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf(stdout.String())
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}

			if test.expected == nil && stdout == nil {
				return
			}

			stdoutStr := stdout.String()
			if !cmp.Equal(stdoutStr, *test.expected) {
				t.Errorf("stdout mismatch (-want +got):\n%s", cmp.Diff(stdoutStr, *test.expected))
			}
		})
	}
}
