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

package evaluate_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
)

type executeWithPiper interface {
	ExecuteWithPipe(command string, value string, timeout int) (*bytes.Buffer, error)
}

type executer struct {
	executeWithPipe func(command string, value string, timeout int) (*bytes.Buffer, error)
}

func (e *executer) ExecuteWithPipe(command string, value string, timeout int) (*bytes.Buffer, error) {
	return e.executeWithPipe(command, value, timeout)
}

func int32ToPtr(value int32) *int32 {
	return &value
}

func TestGetEvaluation(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expectedErr error
		expected    *evaluate.Evaluation
		metrics     *metric.ResourceMetrics
		config      *config.Config
		executer    executeWithPiper
	}{
		{
			"Execute fail",
			errors.New("fail to evaluate"),
			nil,
			&metric.ResourceMetrics{
				Metrics: []*metric.Metric{
					&metric.Metric{
						Resource: "test pod",
						Value:    "test value",
					},
				},
			},
			&config.Config{
				Evaluate:        "test evaluate command",
				EvaluateTimeout: 10,
			},
			func() *executer {
				execute := executer{}
				execute.executeWithPipe = func(command string, value string, timeout int) (*bytes.Buffer, error) {
					return nil, errors.New("fail to evaluate")
				}
				return &execute
			}(),
		},
		{
			"Execute success with valid JSON",
			nil,
			&evaluate.Evaluation{
				TargetReplicas: int32ToPtr(int32(3)),
			},
			&metric.ResourceMetrics{
				Metrics: []*metric.Metric{
					&metric.Metric{
						Resource: "test pod",
						Value:    "test value",
					},
				},
			},
			&config.Config{
				Evaluate:        "test evaluate command",
				EvaluateTimeout: 10,
			},
			func() *executer {
				execute := executer{}
				execute.executeWithPipe = func(command string, value string, timeout int) (*bytes.Buffer, error) {
					// Convert into JSON
					jsonEvaluation, err := json.Marshal(&evaluate.Evaluation{
						TargetReplicas: int32ToPtr(int32(3)),
					})
					if err != nil {
						return nil, err
					}
					var buffer bytes.Buffer
					buffer.WriteString(string(jsonEvaluation))
					return &buffer, nil
				}
				return &execute
			}(),
		},
		{
			"Execute success with invalid evaluation",
			&evaluate.ErrInvalidEvaluation{
				Message: `Invalid evaluation returned by evaluator: { "invalid": "invalid"}`,
			},
			nil,
			&metric.ResourceMetrics{
				Metrics: []*metric.Metric{
					&metric.Metric{
						Resource: "test pod",
						Value:    "test value",
					},
				},
			},
			&config.Config{
				Evaluate:        "test evaluate command",
				EvaluateTimeout: 10,
			},
			func() *executer {
				execute := executer{}
				execute.executeWithPipe = func(command string, value string, timeout int) (*bytes.Buffer, error) {
					var buffer bytes.Buffer
					buffer.WriteString(`{ "invalid": "invalid"}`)
					return &buffer, nil
				}
				return &execute
			}(),
		},
		{
			"Execute success with invalid JSON",
			errors.New(`invalid character 'i' looking for beginning of value`),
			nil,
			&metric.ResourceMetrics{
				Metrics: []*metric.Metric{
					&metric.Metric{
						Resource: "test pod",
						Value:    "test value",
					},
				},
			},
			&config.Config{
				Evaluate:        "test evaluate command",
				EvaluateTimeout: 10,
			},
			func() *executer {
				execute := executer{}
				execute.executeWithPipe = func(command string, value string, timeout int) (*bytes.Buffer, error) {
					var buffer bytes.Buffer
					buffer.WriteString(`invalid`)
					return &buffer, nil
				}
				return &execute
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			evaluator := &evaluate.Evaluator{
				Config:   test.config,
				Executer: test.executer,
			}
			evaluation, err := evaluator.GetEvaluation(test.metrics)
			if !cmp.Equal(&test.expectedErr, &err, equateErrorMessage) {
				t.Errorf("Error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, evaluation) {
				t.Errorf("Evaluation mismatch (-want +got):\n%s", cmp.Diff(test.expected, evaluation))
			}
		})
	}
}
