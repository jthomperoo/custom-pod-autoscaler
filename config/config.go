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

// Package config handles parsing YAML and environment variable input into a
// Custom Pod Autoscaler configuration file. Contains a set of defaults that
// can be overridden by provided YAML and env vars.
package config

import (
	"encoding/json"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v2"
	autoscaling "k8s.io/api/autoscaling/v1"
)

const (
	// PerPodRunMode runs metric gathering per Pod, individually running the script for each Pod being managed
	// with the Pod information piped into the metric gathering script
	PerPodRunMode = "per-pod"
	// PerResourceRunMode runs metric gathering per Deployment, running the script only once for the resource
	// being managed, with the resource information piped into the metric gathering script
	PerResourceRunMode = "per-resource"
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
	defaultRunMode         = PerPodRunMode
)

const yamlStructTag = "yaml"

// Config is the configuration options for the CPA
type Config struct {
	ScaleTargetRef  *autoscaling.CrossVersionObjectReference `yaml:"scaleTargetRef"`
	Evaluate        string                                   `yaml:"evaluate"`
	Metric          string                                   `yaml:"metric"`
	Interval        int                                      `yaml:"interval"`
	Host            string                                   `yaml:"host"`
	Port            int                                      `yaml:"port"`
	EvaluateTimeout int                                      `yaml:"evaluateTimeout"`
	MetricTimeout   int                                      `yaml:"metricTimeout"`
	Namespace       string                                   `yaml:"namespace"`
	MinReplicas     int32                                    `yaml:"minReplicas"`
	MaxReplicas     int32                                    `yaml:"maxReplicas"`
	RunMode         string                                   `yaml:"runMode"`
}

// LoadConfig loads in the default configuration, then overrides it from the config file,
// then any env vars set.
func LoadConfig(configFileData []byte, envVars map[string]string) (*Config, error) {
	config := newDefaultConfig()
	err := loadFromYAML(configFileData, config)
	if err != nil {
		return nil, err
	}
	err = loadFromEnv(config, envVars)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func loadFromYAML(data []byte, config *Config) error {
	err := yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}
	return nil
}

func loadFromEnv(config *Config, envVars map[string]string) error {
	// Get config Go types and values
	configTypes := reflect.TypeOf(config).Elem()
	configValues := reflect.ValueOf(config).Elem()

	// Iterate through each field in the config
	for i := 0; i < configTypes.NumField(); i++ {
		// Get each field's type and value
		fieldType := configTypes.Field(i)
		fieldValue := configValues.Field(i)

		// Extract YAML tag from the type, e.g `yaml:"example"` would return example
		tag := fieldType.Tag.Get(yamlStructTag)

		// Check if there is an environment variable provided with the same tag
		value, exists := envVars[tag]
		if !exists {
			continue
		}

		// Assign values using correct types
		if fieldValue.Kind() == reflect.String {
			fieldValue.SetString(value)
			continue
		}
		if fieldValue.Kind() == reflect.Int {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			fieldValue.SetInt(intVal)
			continue
		}

		// If the type is not one of the primitives above, it must be in JSON form, so try to parse
		// it and set the value from the unmarshalled JSON value
		fieldRef := reflect.New(fieldType.Type)
		err := json.Unmarshal([]byte(value), fieldRef.Interface())
		if err != nil {
			return err
		}

		fieldValue.Set(fieldRef.Elem())
		continue
	}
	return nil
}

func newDefaultConfig() *Config {
	return &Config{
		Interval:        defaultInterval,
		Metric:          defaultMetric,
		Evaluate:        defaultEvaluate,
		Host:            defaultHost,
		Port:            defaultPort,
		EvaluateTimeout: defaultEvaluateTimeout,
		MetricTimeout:   defaultMetricTimeout,
		Namespace:       defaultNamespace,
		MinReplicas:     defaultMinReplicas,
		MaxReplicas:     defaultMaxReplicas,
		RunMode:         defaultRunMode,
	}
}
