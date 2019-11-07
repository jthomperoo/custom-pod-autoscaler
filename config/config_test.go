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

package config_test

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/test"
	"gopkg.in/yaml.v2"
)

const (
	evaluateEnvName        = "evaluate"
	metricEnvName          = "metric"
	intervalEnvName        = "interval"
	hostEnvName            = "host"
	portEnvName            = "port"
	metricTimeoutEnvName   = "metricTimeout"
	evaluateTimeoutEnvName = "evaluateTimeout"
	namespaceEnvName       = "namespace"
	scaleTargetRefEnvName  = "scaleTargetRef"

	defaultEvaluate        = ">&2 echo 'ERROR: No evaluate command set' && exit 1"
	defaultMetric          = ">&2 echo 'ERROR: No metric command set' && exit 1"
	defaultInterval        = 15000
	defaultHost            = "0.0.0.0"
	defaultPort            = 5000
	defaultMetricTimeout   = 5000
	defaultEvaluateTimeout = 5000
	defaultNamespace       = "default"

	invalidYAML                   = "- in: -: valid - yaml"
	testEvaluate                  = "test evaluate"
	testMetric                    = "test metric"
	testInterval                  = 1234
	testHost                      = "1.2.3.4"
	testPort                      = 1234
	testMetricTimeout             = 4321
	testEvaluateTimeout           = 8765
	testNamespace                 = "test namespace"
	testScaleTargetRefKind        = "test kind"
	testScaleTargetRefName        = "test name"
	testScaleTargetRefAPIVersion  = "test api version"
	testScaleTargetRefJSON        = "{\"kind\":\"test kind\", \"name\":\"test name\", \"apiVersion\":\"test api version\" }"
	testScaleTargetRefInvalidJSON = "abc invali:d json"
)

func TestLoadConfig_InvalidYAML(t *testing.T) {
	_, err := config.LoadConfig([]byte(invalidYAML), nil)
	if err == nil {
		t.Errorf("Expected error due to invalid YAML provided")
	}
}

func TestLoadConfig_InvalidIntEnv(t *testing.T) {
	_, err := config.LoadConfig(nil, map[string]string{
		intervalEnvName: "invalid",
	})

	if err == nil {
		t.Errorf("Expected error due to invalid integer environment variable provided")
	}
}

func TestLoadConfig_NoYAML(t *testing.T) {
	testConfig := test.GetTestConfig()

	testEnvVars := getTestEnvVars()

	loadedConfig, err := config.LoadConfig(nil, testEnvVars)
	if err != nil {
		t.Error(err)
	}

	if !cmp.Equal(loadedConfig, testConfig) {
		t.Errorf("Config mismatch (-want +got):\n%s", cmp.Diff(testConfig, loadedConfig))
	}
}

func TestLoadConfig_NoEnv(t *testing.T) {
	testConfig := test.GetTestConfig()

	yamlConfig, err := yaml.Marshal(testConfig)
	if err != nil {
		t.Error(err)
	}

	loadedConfig, err := config.LoadConfig(yamlConfig, nil)
	if err != nil {
		t.Error(err)
	}

	if !cmp.Equal(loadedConfig, testConfig) {
		t.Errorf("Config mismatch (-want +got):\n%s", cmp.Diff(testConfig, loadedConfig))
	}
}

func TestLoadConfig_NoYAMLNoEnv(t *testing.T) {
	testConfig := getDefaultConfig()

	loadedConfig, err := config.LoadConfig(nil, nil)
	if err != nil {
		t.Error(err)
	}

	if !cmp.Equal(loadedConfig, testConfig) {
		t.Errorf("Config mismatch (-want +got):\n%s", cmp.Diff(testConfig, loadedConfig))
	}
}

func TestLoadCOnfig_InvalidScaleTargetRefJSON(t *testing.T) {
	testEnvVars := getTestEnvVars()

	testEnvVars[scaleTargetRefEnvName] = testScaleTargetRefInvalidJSON

	_, err := config.LoadConfig(nil, testEnvVars)
	if err == nil {
		t.Errorf("Expected error due to invalid scaleTargetRef environment variable provided")
	}
}

func getDefaultConfig() *config.Config {
	return &config.Config{
		Evaluate:        defaultEvaluate,
		Metric:          defaultMetric,
		Interval:        defaultInterval,
		Host:            defaultHost,
		Port:            defaultPort,
		MetricTimeout:   defaultMetricTimeout,
		EvaluateTimeout: defaultEvaluateTimeout,
		Namespace:       defaultNamespace,
		ScaleTargetRef:  nil,
	}
}

func getTestEnvVars() map[string]string {
	return map[string]string{
		evaluateEnvName:        testEvaluate,
		metricEnvName:          testMetric,
		intervalEnvName:        strconv.FormatInt(testInterval, 10),
		hostEnvName:            testHost,
		portEnvName:            strconv.FormatInt(testPort, 10),
		metricTimeoutEnvName:   strconv.FormatInt(testMetricTimeout, 10),
		evaluateTimeoutEnvName: strconv.FormatInt(testEvaluateTimeout, 10),
		namespaceEnvName:       testNamespace,
		scaleTargetRefEnvName:  testScaleTargetRefJSON,
	}
}
