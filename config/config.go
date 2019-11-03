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

package config

import (
	"reflect"
	"strconv"

	"gopkg.in/yaml.v2"
)

const (
	defaultEvaluate        = ">&2 echo 'ERROR: No evaluate command set' && exit 1"
	defaultMetric          = ">&2 echo 'ERROR: No metric command set' && exit 1"
	defaultInterval        = 15000
	defaultSelector        = ""
	defaultHost            = "0.0.0.0"
	defaultPort            = 5000
	defaultMetricTimeout   = 5000
	defaultEvaluateTimeout = 5000
	defaultNamespace       = "default"
)

const yamlStructTag = "yaml"

// Config is the configuration options for the CPA
type Config struct {
	Evaluate        string `yaml:"evaluate"`
	Metric          string `yaml:"metric"`
	Interval        int    `yaml:"interval"`
	Selector        string `yaml:"selector"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	EvaluateTimeout int    `yaml:"evaluate_timeout"`
	MetricTimeout   int    `yaml:"metric_timeout"`
	Namespace       string `yaml:"namespace"`
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
		}
		if fieldValue.Kind() == reflect.Int {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			fieldValue.SetInt(intVal)
		}
	}
	return nil
}

func newDefaultConfig() *Config {
	return &Config{
		Interval:        defaultInterval,
		Metric:          defaultMetric,
		Evaluate:        defaultEvaluate,
		Selector:        defaultSelector,
		Host:            defaultHost,
		Port:            defaultPort,
		EvaluateTimeout: defaultEvaluateTimeout,
		MetricTimeout:   defaultMetricTimeout,
		Namespace:       defaultNamespace,
	}
}
