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

// Package shell handles interactions with the OS shell
package shell

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
)

// Type shell represents a shell command
const Type = "shell"

// Command represents the function that builds the exec.Cmd to be used in shell commands.
type command = func(name string, arg ...string) *exec.Cmd

// Execute represents a way to execute shell commands with values piped to them.
type Execute struct {
	Command command
}

// ExecuteWithValue executes a shell command with a value piped to it.
// If it exits with code 0, no error is returned and the stdout is captured and returned.
// If it exits with code 1, an error is returned and the stderr is captured and returned.
// If the timeout is reached, an error is returned.
func (e *Execute) ExecuteWithValue(method *config.Method, value string) (string, error) {
	// Build command string with value piped into it
	cmd := e.Command(method.Shell.Entrypoint, method.Shell.Command)

	// Set up byte buffer to write values to stdin
	inb := bytes.Buffer{}
	_, err := inb.WriteString(value)
	if err != nil {
		return "", err
	}
	cmd.Stdin = &inb

	// Set up byte buffers to read stdout and stderr
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	// Start command
	err = cmd.Start()
	if err != nil {
		return "", err
	}

	// Set up channel to wait for command to finish
	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	// Set up a timeout, after which if the command hasn't finished it will be stopped
	timeoutListener := time.After(time.Duration(method.Timeout) * time.Millisecond)

	select {
	case <-timeoutListener:
		cmd.Process.Kill()
		return "", fmt.Errorf("Entrypoint '%s', command '%s' timed out", method.Shell.Entrypoint, method.Shell.Command)
	case err = <-done:
		if err != nil {
			fmt.Println(fmt.Sprintf("stderr: %s", errb.String()))
			return "", err
		}
	}
	return outb.String(), nil
}

// GetType returns the shell executer type
func (e *Execute) GetType() string {
	return Type
}
