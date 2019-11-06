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

package shell_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/shell"
)

const (
	testPipeValue = "test pipe value"
	testStdout    = "test std out"
	testStderr    = "test std err"
	testCommand   = "command"
	testTimeout   = 100
)

func TestExecWithValuePipe_Success(t *testing.T) {
	s := shell.NewCommandExecuteWithPipe(fakeExecCommandSuccess)
	stdout, err := s.ExecuteWithPipe(testCommand, testPipeValue, testTimeout)
	if err != nil {
		t.Error(stdout.String())
		t.Error(err)
		return
	}

	stdoutStr := stdout.String()
	if !cmp.Equal(stdoutStr, testStdout) {
		t.Errorf("stdout mismatch (-want +got):\n%s", cmp.Diff(testStdout, stdoutStr))
	}
}

func TestExecWithValuePipe_Fail(t *testing.T) {
	s := shell.NewCommandExecuteWithPipe(fakeExecCommandFailure)
	stderr, err := s.ExecuteWithPipe(testCommand, testPipeValue, testTimeout)
	if err == nil {
		t.Errorf("Expected error due to shell command exiting with non-zero exit code")
		return
	}

	stderrStr := stderr.String()
	if !cmp.Equal(stderrStr, testStderr) {
		t.Errorf("stderr mismatch (-want +got):\n%s", cmp.Diff(testStderr, stderrStr))
	}
}

func TestExecWithValuePipe_Timeout(t *testing.T) {
	s := shell.NewCommandExecuteWithPipe(fakeExecCommandTimeout)
	stderr, err := s.ExecuteWithPipe(testCommand, testPipeValue, testTimeout)
	if err == nil {
		t.Errorf("Expected error due to shell command timing out")
		return
	}

	if stderr != nil {
		t.Errorf("Expected no stderr due to shell command timing out")
	}
}

func TestExecWithValuePipe_StartFail(t *testing.T) {
	s := shell.NewCommandExecuteWithPipe(fakeExecCommandStartFail)
	stderr, err := s.ExecuteWithPipe(testCommand, testPipeValue, testTimeout)
	if err == nil {
		t.Errorf("Expected error due to shell failing to start")
		return
	}

	if stderr != nil {
		t.Errorf("Expected no stderr due to shell command failing to start")
	}
}

// Test shell commands

func TestShellProcessSuccess(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") != "1" {
		return
	}

	// Due to the fake shell command, the actual command and value piped to it
	// are arguments passed to this command - at argument position 5
	// e.g. echo 'stdin' | command
	commandAndPipe := os.Args[5]

	stdin := strings.TrimSpace(strings.Split(commandAndPipe, "|")[0])
	command := strings.TrimSpace(strings.Split(commandAndPipe, "|")[1])

	// stdin is echoed and piped to the command, so the argument will be surrounded
	// by an echo command
	testPipeValueWithEcho := fmt.Sprintf("echo '%s'", testPipeValue)

	// Check command is correct
	if !cmp.Equal(command, testCommand) {
		fmt.Fprintf(os.Stderr, "stdin mismatch (-want +got):\n%s", cmp.Diff(testCommand, command))
		os.Exit(1)
	}

	// Check piped value in is correct
	if !cmp.Equal(stdin, testPipeValueWithEcho) {
		fmt.Fprintf(os.Stderr, "stdin mismatch (-want +got):\n%s", cmp.Diff(testPipeValueWithEcho, stdin))
		os.Exit(1)
	}

	fmt.Fprint(os.Stdout, testStdout)
	os.Exit(0)
}

func TestShellProcessFail(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stderr, testStderr)
	os.Exit(1)
}

func TestShellProcessTimeout(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") != "1" {
		return
	}

	time.Sleep(testTimeout * 2 * time.Millisecond)
	os.Exit(0)
}

func fakeExecCommandSuccess(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestShellProcessSuccess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	return cmd
}

func fakeExecCommandFailure(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestShellProcessFail", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	return cmd
}

func fakeExecCommandTimeout(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestShellProcessTimeout", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	return cmd
}

func fakeExecCommandStartFail(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestShellProcessSuccess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	cmd.Start()
	return cmd
}
