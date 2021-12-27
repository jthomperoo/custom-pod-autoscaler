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

package confload_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/confload"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
)

const (
	defaultInterval                = 15000
	defaultMetricTimeout           = 5000
	defaultEvaluateTimeout         = 5000
	defaultNamespace               = "default"
	defaultMinReplicas             = 1
	defaultMaxReplicas             = 10
	defaultStartTime               = 1
	defaultRunMode                 = "per-pod"
	defaultLogVerbosity            = 0
	defaultDownscaleStabilization  = 0
	defaultCPUInitializationPeriod = 300
	defaultInitialReadinessDelay   = 30
)

const (
	defaultAPIEnabled = true
	defaultHost       = "0.0.0.0"
	defaultPort       = 5000
	defaultUseHTTPS   = false
	defaultCertFile   = ""
	defaultKeyFile    = ""
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
			errors.New("error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type v2beta2.CrossVersionObjectReference"),
			nil,
		},
		{
			"No JSON or YAML no env return default",
			nil,
			nil,
			nil,
			&config.Config{
				Interval:               defaultInterval,
				Namespace:              defaultNamespace,
				RunMode:                defaultRunMode,
				MinReplicas:            defaultMinReplicas,
				MaxReplicas:            defaultMaxReplicas,
				StartTime:              defaultStartTime,
				LogVerbosity:           defaultLogVerbosity,
				DownscaleStabilization: defaultDownscaleStabilization,
				APIConfig: &config.APIConfig{
					Enabled:  defaultAPIEnabled,
					UseHTTPS: defaultUseHTTPS,
					Port:     defaultPort,
					Host:     defaultHost,
					CertFile: defaultCertFile,
					KeyFile:  defaultKeyFile,
				},
				ScaleTargetRef:          nil,
				InitialReadinessDelay:   defaultInitialReadinessDelay,
				CPUInitializationPeriod: defaultCPUInitializationPeriod,
			},
		},
		{
			"No JSON or YAML override with env",
			nil,
			map[string]string{
				"metric": `{
					"type" : "shell",
					"timeout": 10,
					"shell": {
						"command" : ["testcommand"],
						"entrypoint" : "testentry"
					}
				}`,
				"interval":               "35",
				"namespace":              "test env namespace",
				"runMode":                "test run mode",
				"minReplicas":            "3",
				"maxReplicas":            "6",
				"startTime":              "0",
				"scaleTargetRef":         `{ "name": "test target name", "kind": "test target kind", "apiVersion": "test target api version"}`,
				"logVerbosity":           "3",
				"downscaleStabilization": "200",
				"apiConfig": `{
					"port": 3000,
					"enabled": true,
					"useHTTPS": true,
					"certFile": "testcert",
					"keyFile": "testkey"
				}`,
			},
			nil,
			&config.Config{
				Interval:    35,
				Namespace:   "test env namespace",
				RunMode:     "test run mode",
				MinReplicas: 3,
				MaxReplicas: 6,
				StartTime:   0,
				Metric: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"testcommand"},
						Entrypoint: "testentry",
					},
				},
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Name:       "test target name",
					Kind:       "test target kind",
					APIVersion: "test target api version",
				},
				LogVerbosity:           3,
				DownscaleStabilization: 200,
				APIConfig: &config.APIConfig{
					Enabled:  true,
					UseHTTPS: true,
					Port:     3000,
					CertFile: "testcert",
					KeyFile:  "testkey",
				},
				InitialReadinessDelay:   defaultInitialReadinessDelay,
				CPUInitializationPeriod: defaultCPUInitializationPeriod,
			},
		},
		{
			"No env override with YAML",
			[]byte(strings.Replace(`
				evaluate:
				  type: shell
				  timeout: 10
				  shell:
				    command:
				      - "testcommand"
				      - "arg1"
				    entrypoint: "testentry"
				metric:
				  type: http
				  timeout: 10
				  http:
				    method: "GET"
				    url: "https://www.custompodautoscaler.com"
				    successCodes:
				      - 200
				    headers:
				      a: testa
				      b: testb
				      c: testc
				    parameterMode: query
				interval: 30
				runMode: "test run mode"
				minReplicas: 2
				maxReplicas: 7
				startTime: 0
				namespace: "test yaml namespace"
				logVerbosity: 2
				downscaleStabilization: 200
				apiConfig:
				  enabled: true
				  useHTTPS: false
				  host: "test yaml host"
				  port: 7890
			`, "\t", "", -1)),
			nil,
			nil,
			&config.Config{
				Interval:    30,
				RunMode:     "test run mode",
				MinReplicas: 2,
				MaxReplicas: 7,
				StartTime:   0,
				Namespace:   "test yaml namespace",
				Evaluate: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"testcommand", "arg1"},
						Entrypoint: "testentry",
					},
				},
				Metric: &config.Method{
					Type:    "http",
					Timeout: 10,
					HTTP: &config.HTTP{
						Method:        "GET",
						URL:           "https://www.custompodautoscaler.com",
						SuccessCodes:  []int{200},
						ParameterMode: "query",
						Headers: map[string]string{
							"a": "testa",
							"b": "testb",
							"c": "testc",
						},
					},
				},
				LogVerbosity:           2,
				DownscaleStabilization: 200,
				APIConfig: &config.APIConfig{
					Enabled:  true,
					UseHTTPS: false,
					Port:     7890,
					Host:     "test yaml host",
				},
				InitialReadinessDelay:   defaultInitialReadinessDelay,
				CPUInitializationPeriod: defaultCPUInitializationPeriod,
			},
		},
		{
			"No env override with JSON",
			[]byte(`{
				"evaluate":{
					"type":"shell",
					"timeout":10,
					"shell": {
						"command": ["testcommand", "arg1"],
						"entrypoint": "testentry"
					}
				},
				"metric":{
					"type":"http",
					"timeout":10,
					"http": {
						"method": "POST",
						"url": "https://www.custompodautoscaler.com",
						"successCodes": [
							200
						],
						"headers": {
							"a": "testa",
							"b": "testb",
							"c": "testc"
						},
						"parameterMode": "body"
					}
				},
				"interval":30,
				"runMode":"test run mode",
				"minReplicas":2,
				"maxReplicas":7,
				"startTime":0,
				"namespace":"test yaml namespace",
				"logVerbosity":1,
				"downscaleStabilization":200,
				"apiConfig": {
					"port": 7890,
					"enabled": false,
					"useHTTPS": true,
					"certFile": "test cert file",
					"keyFile": "test key file"
				}
			}`),
			nil,
			nil,
			&config.Config{
				Interval:    30,
				RunMode:     "test run mode",
				MinReplicas: 2,
				MaxReplicas: 7,
				StartTime:   0,
				Namespace:   "test yaml namespace",
				Evaluate: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"testcommand", "arg1"},
						Entrypoint: "testentry",
					},
				},
				Metric: &config.Method{
					Type:    "http",
					Timeout: 10,
					HTTP: &config.HTTP{
						Method:        "POST",
						URL:           "https://www.custompodautoscaler.com",
						SuccessCodes:  []int{200},
						ParameterMode: "body",
						Headers: map[string]string{
							"a": "testa",
							"b": "testb",
							"c": "testc",
						},
					},
				},
				LogVerbosity:           1,
				DownscaleStabilization: 200,
				APIConfig: &config.APIConfig{
					Enabled:  false,
					UseHTTPS: true,
					Port:     7890,
					Host:     "0.0.0.0",
					CertFile: "test cert file",
					KeyFile:  "test key file",
				},
				InitialReadinessDelay:   defaultInitialReadinessDelay,
				CPUInitializationPeriod: defaultCPUInitializationPeriod,
			},
		},
		{
			"Partial YAML partial env",
			[]byte(strings.Replace(`
				evaluate:
				  type: shell
				  timeout: 10
				  shell:
				    command:
				      - "testcommand"
				      - "arg1"
				    entrypoint: "testentry"
				metric:
				  type: shell
				  timeout: 10
				  shell:
				    command:
				      - "testcommand"
				    entrypoint: "testentry"
				apiConfig:
				  enabled: true
				  useHTTPS: false
				  host: "test host"
				  port: 7890
				runMode: "test run mode"
				namespace: "test yaml namespace"
			`, "\t", "", -1)),
			map[string]string{
				"interval":               "35",
				"minReplicas":            "3",
				"maxReplicas":            "6",
				"startTime":              "0",
				"logVerbosity":           "5",
				"downscaleStabilization": "300",
				"scaleTargetRef":         `{ "name": "test target name", "kind": "test target kind", "apiVersion": "test target api version"}`,
			},
			nil,
			&config.Config{
				Interval:    35,
				RunMode:     "test run mode",
				MinReplicas: 3,
				MaxReplicas: 6,
				StartTime:   0,
				Namespace:   "test yaml namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Name:       "test target name",
					Kind:       "test target kind",
					APIVersion: "test target api version",
				},
				Evaluate: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"testcommand", "arg1"},
						Entrypoint: "testentry",
					}},
				Metric: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"testcommand"},
						Entrypoint: "testentry",
					},
				},
				LogVerbosity:           5,
				DownscaleStabilization: 300,
				APIConfig: &config.APIConfig{
					Enabled:  true,
					UseHTTPS: false,
					Port:     7890,
					Host:     "test host",
				},
				InitialReadinessDelay:   defaultInitialReadinessDelay,
				CPUInitializationPeriod: defaultCPUInitializationPeriod,
			},
		},
		{
			"Partial JSON partial env",
			[]byte(`{
				"evaluate": {
					"type": "shell",
					"timeout": 10,
					"shell": {
						"command": ["testcommand", "arg1"],
						"entrypoint": "testentry"
					}
				},
				"metric": {
					"type": "shell",
					"timeout": 10,
					"shell": {
						"command": ["testcommand"],
						"entrypoint": "testentry"
					}
				},
				"host": "test yaml host",
				"port": 7890,
				"runMode": "test run mode",
				"namespace": "test yaml namespace"
			}`),
			map[string]string{
				"interval":               "35",
				"minReplicas":            "3",
				"maxReplicas":            "6",
				"startTime":              "0",
				"scaleTargetRef":         `{ "name": "test target name", "kind": "test target kind", "apiVersion": "test target api version"}`,
				"logVerbosity":           "3",
				"downscaleStabilization": "300",
				"apiConfig": strings.Replace(`
				host: "test host"
				port: 7890
				enabled: true
				useHTTPS: false
				`, "\t", "", -1),
			},
			nil,
			&config.Config{
				Interval:    35,
				RunMode:     "test run mode",
				MinReplicas: 3,
				MaxReplicas: 6,
				StartTime:   0,
				Namespace:   "test yaml namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Name:       "test target name",
					Kind:       "test target kind",
					APIVersion: "test target api version",
				},
				Evaluate: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"testcommand", "arg1"},
						Entrypoint: "testentry",
					},
				},
				Metric: &config.Method{
					Type:    "shell",
					Timeout: 10,
					Shell: &config.Shell{
						Command:    []string{"testcommand"},
						Entrypoint: "testentry",
					},
				},
				LogVerbosity:           3,
				DownscaleStabilization: 300,
				APIConfig: &config.APIConfig{
					Enabled:  true,
					UseHTTPS: false,
					Port:     7890,
					Host:     "test host",
				},
				InitialReadinessDelay:   defaultInitialReadinessDelay,
				CPUInitializationPeriod: defaultCPUInitializationPeriod,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			config, err := confload.Load(test.yamlInput, test.envVars)
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
