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
	"log"
	"os/exec"
	"time"
)

// ExecWithValuePipe executes a shell command with a value piped to it
func ExecWithValuePipe(command string, value string, timeout int) (*bytes.Buffer, error) {
	// Build command string with value piped into it
	commandString := fmt.Sprintf("echo '%s' | %s", value, command)
	cmd := exec.Command("/bin/sh", "-c", commandString)

	// Set up byte buffers to read stdout and stderr
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	// Start command
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	// Set up channel to wait for command to finish
	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	// Set up a timeout, after which if the command hasn't finished it will be stopped
	timeoutListener := time.After(time.Duration(timeout) * time.Millisecond)

	select {
	case <-timeoutListener:
		cmd.Process.Kill()
		return nil, fmt.Errorf("Command %s timed out", command)
	case err = <-done:
		if err != nil {
			// Output stderr
			log.Println(errb.String())
			return nil, err
		}
	}
	return &outb, nil
}
