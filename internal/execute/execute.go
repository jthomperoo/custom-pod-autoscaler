/*
Copyright 2021 The Custom Pod Autoscaler Authors.

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

// Package execute abstracts methods, providing a standardised way to trigger methods
// and provide values
package execute

import (
	"fmt"

	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
)

// Executer interface provides methods for executing user logic with a value passed through to it
type Executer interface {
	ExecuteWithValue(method *config.Method, value string) (string, error)
	GetType() string
}

// CombinedType is the type of the CombinedExecute; designed to link together multiple executers
// and to provide a simplified single entry point
const CombinedType = "combined"

// CombinedExecute is an executer that contains subexecuters that it will forward method requests
// to; designed to link together multiple executers and to provide a simplified single entry point
type CombinedExecute struct {
	Executers []Executer
}

// ExecuteWithValue takes in a method and a value to pass, it will look at the stored sub executers
// and decide which executer to use for the method provided
func (e *CombinedExecute) ExecuteWithValue(method *config.Method, value string) (string, error) {
	for _, executer := range e.Executers {
		if executer.GetType() == method.Type {
			gathered, err := executer.ExecuteWithValue(method, value)
			if err != nil {
				return "", err
			}
			return gathered, nil
		}
	}
	return "", fmt.Errorf("Unknown execution method: '%s'", method.Type)
}

// GetType returns the CombinedExecute type
func (e *CombinedExecute) GetType() string {
	return CombinedType
}
