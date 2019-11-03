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
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	defaultEvaluate  = ">&2 echo 'ERROR: No evaluate command set' && exit 1"
	defaultMetric    = ">&2 echo 'ERROR: No metric command set' && exit 1"
	defaultInterval  = 15000
	defaultSelector  = ""
	defaultHost      = "0.0.0.0"
	defaultPort      = 5000
	defaultNamespace = "default"
)

// Config is the configuration options for the CPA
type Config struct {
	Evaluate  string `yaml:"evaluate"`
	Metric    string `yaml:"metric"`
	Interval  int    `yaml:"interval"`
	Selector  string `yaml:"selector"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Namespace string `yaml:"-"`
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
	ps := reflect.ValueOf(config)
	s := ps.Elem()
	for key, value := range envVars {
		field := s.FieldByName(strings.Title(key))
		if !field.IsValid() || !field.CanSet() {
			return fmt.Errorf("Field %s is invalid", key)
		}
		if field.Kind() == reflect.String {
			field.SetString(value)
		}
		if field.Kind() == reflect.Int {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(intVal)
		}
	}
	return nil
}

func newDefaultConfig() *Config {
	return &Config{
		Interval:  defaultInterval,
		Metric:    defaultMetric,
		Evaluate:  defaultEvaluate,
		Selector:  defaultSelector,
		Host:      defaultHost,
		Port:      defaultPort,
		Namespace: defaultNamespace,
	}
}
