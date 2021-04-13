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

package config

// Method describes a method for passing data/triggerering logic, such as through a shell
// command
type Method struct {
	Type    string `json:"type"`
	Timeout int    `json:"timeout"`
	Shell   *Shell `json:"shell"`
	HTTP    *HTTP  `json:"http"`
}

// Shell describes configuration options for a shell command method
type Shell struct {
	Command    []string `json:"command"`
	Entrypoint string   `json:"entrypoint"`
}

// HTTP describes configuration options for an HTTP request method
type HTTP struct {
	Method        string            `json:"method"`
	URL           string            `json:"url"`
	Headers       map[string]string `json:"headers,omitempty"`
	SuccessCodes  []int             `json:"successCodes"`
	ParameterMode string            `json:"parameterMode"`
}
