//go:build unit
// +build unit

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

package evaluatecalc_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/evaluatecalc"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/execute"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/metric"
)

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
		info        evaluate.Info
		config      *config.Config
		execute     execute.Executer
	}{
		{
			"Pre-evaluate hook fail",
			errors.New("pre-evaluate hook fail"),
			nil,
			evaluate.Info{
				Metrics: []*metric.ResourceMetric{
					{
						Resource: "test pod",
						Value:    "test value",
					},
				},
				RunType: config.ScalerRunType,
			},
			&config.Config{
				PreEvaluate: &config.Method{
					Type: "fake",
				},
			},
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "", errors.New("pre-evaluate hook fail")
				}
				return &execute
			}(),
		},
		{
			"Execute fail",
			errors.New("fail to evaluate"),
			nil,
			evaluate.Info{
				Metrics: []*metric.ResourceMetric{
					{
						Resource: "test pod",
						Value:    "test value",
					},
				},
				RunType: config.ScalerRunType,
			},
			&config.Config{
				Evaluate: &config.Method{
					Type:    "fake",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"test evaluate command"},
						Entrypoint: "testentry",
					},
				},
			},
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "", errors.New("fail to evaluate")
				}
				return &execute
			}(),
		},
		{
			"Post-evaluate hook fail",
			errors.New("post-evaluate hook fail"),
			nil,
			evaluate.Info{
				Metrics: []*metric.ResourceMetric{
					{
						Resource: "test pod",
						Value:    "test value",
					},
				},
				RunType: config.ScalerRunType,
			},
			&config.Config{
				Evaluate: &config.Method{
					Type:    "execute",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"test evaluate command"},
						Entrypoint: "testentry",
					},
				},
				PostEvaluate: &config.Method{
					Type: "postEvaluate",
				},
			},
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					if method.Type == "postEvaluate" {
						return "", errors.New("post-evaluate hook fail")
					}
					return `{ "targetReplicas" : 3 }`, nil
				}
				return &execute
			}(),
		},
		{
			"Execute success with valid JSON, run pre-evaluate hook",
			nil,
			&evaluate.Evaluation{
				TargetReplicas: int32(3),
			},
			evaluate.Info{
				Metrics: []*metric.ResourceMetric{
					{
						Resource: "test pod",
						Value:    "test value",
					},
				},
				RunType: config.ScalerRunType,
			},
			&config.Config{
				Evaluate: &config.Method{
					Type:    "fake",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"test evaluate command"},
						Entrypoint: "testentry",
					},
				},
				PreEvaluate: &config.Method{
					Type: "fake",
				},
			},
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					// Convert into JSON
					jsonEvaluation, err := json.Marshal(&evaluate.Evaluation{
						TargetReplicas: int32(3),
					})
					if err != nil {
						return "", err
					}
					return string(jsonEvaluation), nil
				}
				return &execute
			}(),
		},
		{
			"Execute success with valid JSON, run post-evaluate hook",
			nil,
			&evaluate.Evaluation{
				TargetReplicas: int32(3),
			},
			evaluate.Info{
				Metrics: []*metric.ResourceMetric{
					{
						Resource: "test pod",
						Value:    "test value",
					},
				},
				RunType: config.ScalerRunType,
			},
			&config.Config{
				Evaluate: &config.Method{
					Type:    "fake",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"test evaluate command"},
						Entrypoint: "testentry",
					},
				},
				PostEvaluate: &config.Method{
					Type: "fake",
				},
			},
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					// Convert into JSON
					jsonEvaluation, err := json.Marshal(&evaluate.Evaluation{
						TargetReplicas: int32(3),
					})
					if err != nil {
						return "", err
					}
					return string(jsonEvaluation), nil
				}
				return &execute
			}(),
		},
		{
			"Execute success with valid JSON",
			nil,
			&evaluate.Evaluation{
				TargetReplicas: int32(3),
			},
			evaluate.Info{
				Metrics: []*metric.ResourceMetric{
					{
						Resource: "test pod",
						Value:    "test value",
					},
				},
				RunType: config.ScalerRunType,
			},
			&config.Config{
				Evaluate: &config.Method{
					Type:    "fake",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"test evaluate command"},
						Entrypoint: "testentry",
					},
				},
			},
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					// Convert into JSON
					jsonEvaluation, err := json.Marshal(&evaluate.Evaluation{
						TargetReplicas: int32(3),
					})
					if err != nil {
						return "", err
					}
					return string(jsonEvaluation), nil
				}
				return &execute
			}(),
		},
		{
			"Execute success with invalid JSON",
			errors.New(`failed to parse JSON evaluation, got 'invalid', err: invalid character 'i' looking for beginning of value`),
			nil,
			evaluate.Info{
				Metrics: []*metric.ResourceMetric{
					{
						Resource: "test pod",
						Value:    "test value",
					},
				},
				RunType: config.ScalerRunType,
			},
			&config.Config{
				Evaluate: &config.Method{
					Type:    "fake",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"test evaluate command"},
						Entrypoint: "testentry",
					},
				},
			},
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "invalid", nil
				}
				return &execute
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			evaluator := &evaluatecalc.Evaluator{
				Config:  test.config,
				Execute: test.execute,
			}
			evaluation, err := evaluator.GetEvaluation(test.info)
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
