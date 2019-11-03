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
	"gopkg.in/yaml.v2"
)

const (
	evaluateEnvName  = "evaluate"
	metricEnvName    = "metric"
	intervalEnvName  = "interval"
	selectorEnvName  = "selector"
	hostEnvName      = "host"
	portEnvName      = "port"
	namespaceEnvName = "namespace"

	defaultEvaluate  = ">&2 echo 'ERROR: No evaluate command set' && exit 1"
	defaultMetric    = ">&2 echo 'ERROR: No metric command set' && exit 1"
	defaultInterval  = 15000
	defaultSelector  = ""
	defaultHost      = "0.0.0.0"
	defaultPort      = 5000
	defaultNamespace = "default"

	invalidYAML   = "- in: -: valid - yaml"
	testEvaluate  = "test evaluate"
	testMetric    = "test metric"
	testInterval  = 1234
	testSelector  = "test selector"
	testHost      = "1.2.3.4"
	testPort      = 1234
	testNamespace = "test namespace"
)

func TestLoadConfig_InvalidYAML(t *testing.T) {
	_, err := config.LoadConfig([]byte(invalidYAML), nil)
	if err == nil {
		t.Errorf("Expected error due to invalid YAML provided")
	}
}

func TestLoadConfig_InvalidEnv(t *testing.T) {
	_, err := config.LoadConfig(nil, map[string]string{
		"invalid": "invalid",
	})

	if err == nil {
		t.Errorf("Expected error due to invalid environment variable provided")
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
	testConfig := &config.Config{
		Evaluate:  testEvaluate,
		Metric:    testMetric,
		Interval:  testInterval,
		Selector:  testSelector,
		Host:      testHost,
		Port:      testPort,
		Namespace: testNamespace,
	}

	testEnvVars := map[string]string{
		evaluateEnvName:  testEvaluate,
		metricEnvName:    testMetric,
		intervalEnvName:  strconv.FormatInt(testInterval, 10),
		selectorEnvName:  testSelector,
		hostEnvName:      testHost,
		portEnvName:      strconv.FormatInt(testPort, 10),
		namespaceEnvName: testNamespace,
	}

	loadedConfig, err := config.LoadConfig(nil, testEnvVars)
	if err != nil {
		t.Error(err)
	}

	if !cmp.Equal(loadedConfig, testConfig) {
		t.Errorf("Config mismatch (-want +got):\n%s", cmp.Diff(testConfig, loadedConfig))
	}
}

func TestLoadConfig_NoEnv(t *testing.T) {
	testConfig := &config.Config{
		Evaluate: testEvaluate,
		Metric:   testMetric,
		Interval: testInterval,
		Selector: testSelector,
		Host:     testHost,
		Port:     testPort,
	}

	yamlConfig, err := yaml.Marshal(testConfig)
	if err != nil {
		t.Error(err)
	}

	loadedConfig, err := config.LoadConfig(yamlConfig, nil)
	if err != nil {
		t.Error(err)
	}

	// Namespace is not loaded in by the YAML, so it should be the default value
	testConfig.Namespace = defaultNamespace

	if !cmp.Equal(loadedConfig, testConfig) {
		t.Errorf("Config mismatch (-want +got):\n%s", cmp.Diff(testConfig, loadedConfig))
	}
}

func TestLoadConfig_NoYAMLNoEnv(t *testing.T) {
	testConfig := &config.Config{
		Evaluate:  defaultEvaluate,
		Metric:    defaultMetric,
		Interval:  defaultInterval,
		Selector:  defaultSelector,
		Host:      defaultHost,
		Port:      defaultPort,
		Namespace: defaultNamespace,
	}

	loadedConfig, err := config.LoadConfig(nil, nil)
	if err != nil {
		t.Error(err)
	}

	if !cmp.Equal(loadedConfig, testConfig) {
		t.Errorf("Config mismatch (-want +got):\n%s", cmp.Diff(testConfig, loadedConfig))
	}
}
