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

package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	v1 "k8s.io/api/autoscaling/v1"
)

const (
	defaultEvaluate        = ">&2 echo 'ERROR: No evaluate command set' && exit 1"
	defaultMetric          = ">&2 echo 'ERROR: No metric command set' && exit 1"
	defaultInterval        = 15000
	defaultHost            = "0.0.0.0"
	defaultPort            = 5000
	defaultMetricTimeout   = 5000
	defaultEvaluateTimeout = 5000
	defaultNamespace       = "default"
	defaultMinReplicas     = 1
	defaultMaxReplicas     = 10
	defaultStartTime       = 1
	defaultRunMode         = "per-pod"
)

func TestLoadConfig(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		yamlInput   []byte
		envVars     map[string]string
		expectedErr error
		expected    *config.Config
	}{
		{
			"Invalid YAML",
			[]byte("invalid"),
			nil,
			errors.New("yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid` into config.Config"),
			nil,
		},
		{
			"Invalid int YAML config",
			[]byte("interval: \"invalid\""),
			nil,
			errors.New("yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid` into int"),
			nil,
		},
		{
			"Invalid int env config",
			nil,
			map[string]string{
				"interval": "invalid",
			},
			errors.New("strconv.ParseInt: parsing \"invalid\": invalid syntax"),
			nil,
		},
		{
			"Invalid scale target ref",
			nil,
			map[string]string{
				"scaleTargetRef": "invalid JSON",
			},
			errors.New("invalid character 'i' looking for beginning of value"),
			nil,
		},
		{
			"No YAML no env return default",
			nil,
			nil,
			nil,
			&config.Config{
				Evaluate:        defaultEvaluate,
				Metric:          defaultMetric,
				Interval:        defaultInterval,
				Host:            defaultHost,
				Port:            defaultPort,
				MetricTimeout:   defaultMetricTimeout,
				EvaluateTimeout: defaultEvaluateTimeout,
				Namespace:       defaultNamespace,
				RunMode:         defaultRunMode,
				MinReplicas:     defaultMinReplicas,
				MaxReplicas:     defaultMaxReplicas,
				StartTime:       defaultStartTime,
				ScaleTargetRef:  nil,
			},
		},
		{
			"No YAML override with env",
			nil,
			map[string]string{
				"evaluate":        "test env evaluate",
				"metric":          "test env metric",
				"interval":        "35",
				"host":            "test env host",
				"port":            "1234",
				"metricTimeout":   "13",
				"evaluateTimeout": "14",
				"namespace":       "test env namespace",
				"runMode":         "test run mode",
				"minReplicas":     "3",
				"maxReplicas":     "6",
				"startTime":       "0",
				"scaleTargetRef":  `{ "name": "test target name", "kind": "test target kind", "apiVersion": "test target api version"}`,
			},
			nil,
			&config.Config{
				Evaluate:        "test env evaluate",
				Metric:          "test env metric",
				Interval:        35,
				Host:            "test env host",
				Port:            1234,
				MetricTimeout:   13,
				EvaluateTimeout: 14,
				Namespace:       "test env namespace",
				RunMode:         "test run mode",
				MinReplicas:     3,
				MaxReplicas:     6,
				StartTime:       0,
				ScaleTargetRef: &v1.CrossVersionObjectReference{
					Name:       "test target name",
					Kind:       "test target kind",
					APIVersion: "test target api version",
				},
			},
		},
		{
			"No env override with YAML",
			[]byte(strings.Replace(`
				evaluate: "test yaml evaluate"
				metric: "test yaml metric"
				interval: 30
				host: "test yaml host"
				port: 7890
				metricTimeout: 10
				evaluateTimeout: 11
				runMode: "test run mode"
				minReplicas: 2
				maxReplicas: 7
				startTime: 0
				namespace: "test yaml namespace"
			`, "\t", "", -1)),
			nil,
			nil,
			&config.Config{
				Evaluate:        "test yaml evaluate",
				Metric:          "test yaml metric",
				Interval:        30,
				Host:            "test yaml host",
				Port:            7890,
				MetricTimeout:   10,
				EvaluateTimeout: 11,
				RunMode:         "test run mode",
				MinReplicas:     2,
				MaxReplicas:     7,
				StartTime:       0,
				Namespace:       "test yaml namespace",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			config, err := config.LoadConfig(test.yamlInput, test.envVars)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("Error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}

			if !cmp.Equal(config, test.expected) {
				t.Errorf("Config mismatch (-want +got):\n%s", cmp.Diff(test.expected, config))
				return
			}
		})
	}
}
