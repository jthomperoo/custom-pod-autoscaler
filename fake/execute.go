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

package fake

import "github.com/jthomperoo/custom-pod-autoscaler/config"

// Execute (fake) allows inserting logic into an executer for testing
type Execute struct {
	ExecuteWithValueReactor func(method *config.Method, value string) (string, error)
	GetTypeReactor          func() string
}

// ExecuteWithValue calls the fake Execute reactor method provided
func (f *Execute) ExecuteWithValue(method *config.Method, value string) (string, error) {
	return f.ExecuteWithValueReactor(method, value)
}

// GetType calls the fake Execute reactor method provided
func (f *Execute) GetType() string {
	return f.GetTypeReactor()
}
