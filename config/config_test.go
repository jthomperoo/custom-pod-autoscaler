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
			"Invalid JSON or YAML",
			[]byte("invalid"),
			nil,
			errors.New("error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type config.Config"),
			nil,
		},
		{
			"Invalid int JSON or YAML config",
			[]byte("interval: \"invalid\""),
			nil,
			errors.New("error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go struct field Config.interval of type int"),
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
			errors.New("error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type v1.CrossVersionObjectReference"),
			nil,
		},
		{
			"No JSON or YAML no env return default",
			nil,
			nil,
			nil,
			&config.Config{
				Interval:       defaultInterval,
				Host:           defaultHost,
				Port:           defaultPort,
				Namespace:      defaultNamespace,
				RunMode:        defaultRunMode,
				MinReplicas:    defaultMinReplicas,
				MaxReplicas:    defaultMaxReplicas,
				StartTime:      defaultStartTime,
				ScaleTargetRef: nil,
			},
		},
		{
			"No JSON or YAML override with env",
			nil,
			map[string]string{
				"metric":         `{ "type" : "shell", "timeout": 10, "shell": "testcommand"}`,
				"interval":       "35",
				"host":           "test env host",
				"port":           "1234",
				"namespace":      "test env namespace",
				"runMode":        "test run mode",
				"minReplicas":    "3",
				"maxReplicas":    "6",
				"startTime":      "0",
				"scaleTargetRef": `{ "name": "test target name", "kind": "test target kind", "apiVersion": "test target api version"}`,
			},
			nil,
			&config.Config{
				Interval:    35,
				Host:        "test env host",
				Port:        1234,
				Namespace:   "test env namespace",
				RunMode:     "test run mode",
				MinReplicas: 3,
				MaxReplicas: 6,
				StartTime:   0,
				Metric: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell:   "testcommand",
				},
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
				evaluate: 
				  type: shell
				  timeout: 10
				  shell: testcommand
				metric: 
				  type: shell
				  timeout: 10
				  shell: testcommand
				interval: 30
				host: "test yaml host"
				port: 7890
				runMode: "test run mode"
				minReplicas: 2
				maxReplicas: 7
				startTime: 0
				namespace: "test yaml namespace"
			`, "\t", "", -1)),
			nil,
			nil,
			&config.Config{
				Interval:    30,
				Host:        "test yaml host",
				Port:        7890,
				RunMode:     "test run mode",
				MinReplicas: 2,
				MaxReplicas: 7,
				StartTime:   0,
				Namespace:   "test yaml namespace",
				Evaluate: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell:   "testcommand",
				},
				Metric: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell:   "testcommand",
				},
			},
		},
		{
			"No env override with JSON",
			[]byte(`{
				"evaluate":{
					"type":"shell",
					"timeout":10,
					"shell":"testcommand"
				},
				"metric":{
					"type":"shell",
					"timeout":10,
					"shell":"testcommand"
				},
				"interval":30,
				"host":"test yaml host",
				"port":7890,
				"runMode":"test run mode",
				"minReplicas":2,
				"maxReplicas":7,
				"startTime":0,
				"namespace":"test yaml namespace"
			}`),
			nil,
			nil,
			&config.Config{
				Interval:    30,
				Host:        "test yaml host",
				Port:        7890,
				RunMode:     "test run mode",
				MinReplicas: 2,
				MaxReplicas: 7,
				StartTime:   0,
				Namespace:   "test yaml namespace",
				Evaluate: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell:   "testcommand",
				},
				Metric: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell:   "testcommand",
				},
			},
		},
		{
			"Partial YAML partial env",
			[]byte(strings.Replace(`
				evaluate: 
				  type: shell
				  timeout: 10
				  shell: testcommand
				metric: 
				  type: shell
				  timeout: 10
				  shell: testcommand
				host: "test yaml host"
				port: 7890
				runMode: "test run mode"
				namespace: "test yaml namespace"
			`, "\t", "", -1)),
			map[string]string{
				"interval":       "35",
				"minReplicas":    "3",
				"maxReplicas":    "6",
				"startTime":      "0",
				"scaleTargetRef": `{ "name": "test target name", "kind": "test target kind", "apiVersion": "test target api version"}`,
			},
			nil,
			&config.Config{
				Interval:    35,
				Host:        "test yaml host",
				Port:        7890,
				RunMode:     "test run mode",
				MinReplicas: 3,
				MaxReplicas: 6,
				StartTime:   0,
				Namespace:   "test yaml namespace",
				ScaleTargetRef: &v1.CrossVersionObjectReference{
					Name:       "test target name",
					Kind:       "test target kind",
					APIVersion: "test target api version",
				},
				Evaluate: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell:   "testcommand",
				},
				Metric: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell:   "testcommand",
				},
			},
		},
		{
			"Partial JSON partial env",
			[]byte(`{
				"evaluate": {
					"type": "shell",
					"timeout": 10,
					"shell": "testcommand"
				},
				"metric": {
					"type": "shell",
					"timeout": 10,
					"shell": "testcommand"
				},
				"host": "test yaml host",
				"port": 7890,
				"runMode": "test run mode",
				"namespace": "test yaml namespace"
			}`),
			map[string]string{
				"interval":       "35",
				"minReplicas":    "3",
				"maxReplicas":    "6",
				"startTime":      "0",
				"scaleTargetRef": `{ "name": "test target name", "kind": "test target kind", "apiVersion": "test target api version"}`,
			},
			nil,
			&config.Config{
				Interval:    35,
				Host:        "test yaml host",
				Port:        7890,
				RunMode:     "test run mode",
				MinReplicas: 3,
				MaxReplicas: 6,
				StartTime:   0,
				Namespace:   "test yaml namespace",
				ScaleTargetRef: &v1.CrossVersionObjectReference{
					Name:       "test target name",
					Kind:       "test target kind",
					APIVersion: "test target api version",
				},
				Evaluate: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell:   "testcommand",
				},
				Metric: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell:   "testcommand",
				},
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
