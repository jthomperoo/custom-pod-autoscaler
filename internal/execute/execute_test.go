//go:build unit
// +build unit

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

package execute_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/execute"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/fake"
)

func TestCombinedExecute_ExecuteWithValue(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})
	var tests = []struct {
		description string
		expected    string
		expectedErr error
		method      *config.Method
		value       string
		executers   []execute.Executer
	}{
		{
			"Fail, no executers provided",
			"",
			errors.New(`Unknown execution method: 'unknown'`),
			&config.Method{
				Type: "unknown",
			},
			"test",
			[]execute.Executer{},
		},
		{
			"Fail, unknown execution method",
			"",
			errors.New(`Unknown execution method: 'unknown'`),
			&config.Method{
				Type: "unknown",
			},
			"test",
			[]execute.Executer{
				&fake.Execute{
					GetTypeReactor: func() string {
						return "fake"
					},
					ExecuteWithValueReactor: func(method *config.Method, value string) (string, error) {
						return "fake", nil
					},
				},
			},
		},
		{
			"Fail, sub executer fails",
			"",
			errors.New("execute error"),
			&config.Method{
				Type: "test",
			},
			"test",
			[]execute.Executer{
				&fake.Execute{
					GetTypeReactor: func() string {
						return "test"
					},
					ExecuteWithValueReactor: func(method *config.Method, value string) (string, error) {
						return "", errors.New("execute error")
					},
				},
			},
		},
		{
			"Successful execute, one executer",
			"test",
			nil,
			&config.Method{
				Type: "test",
			},
			"test",
			[]execute.Executer{
				&fake.Execute{
					GetTypeReactor: func() string {
						return "test"
					},
					ExecuteWithValueReactor: func(method *config.Method, value string) (string, error) {
						return "test", nil
					},
				},
			},
		},
		{
			"Successful execute, three executers",
			"test",
			nil,
			&config.Method{
				Type: "test1",
			},
			"test",
			[]execute.Executer{
				&fake.Execute{
					GetTypeReactor: func() string {
						return "test1"
					},
					ExecuteWithValueReactor: func(method *config.Method, value string) (string, error) {
						return "test", nil
					},
				},
				&fake.Execute{
					GetTypeReactor: func() string {
						return "test2"
					},
					ExecuteWithValueReactor: func(method *config.Method, value string) (string, error) {
						return "", nil
					},
				},
				&fake.Execute{
					GetTypeReactor: func() string {
						return "test3"
					},
					ExecuteWithValueReactor: func(method *config.Method, value string) (string, error) {
						return "", nil
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			execute := &execute.CombinedExecute{
				Executers: test.executers,
			}
			result, err := execute.ExecuteWithValue(test.method, test.value)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestCombinedExecute_GetType(t *testing.T) {
	var tests = []struct {
		description string
		expected    string
		executers   []execute.Executer
	}{
		{
			"Return type",
			"combined",
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			execute := &execute.CombinedExecute{
				Executers: test.executers,
			}
			result := execute.GetType()
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
