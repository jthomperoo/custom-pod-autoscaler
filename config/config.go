/*
Copyright 2020 The Custom Pod Autoscaler Authors.

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
	"bytes"
	"reflect"
	"strconv"
	"strings"

	autoscaling "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	// PerPodRunMode runs metric gathering per Pod, individually running the script for each Pod being managed
	// with the Pod information piped into the metric gathering script
	PerPodRunMode = "per-pod"
	// PerResourceRunMode runs metric gathering per Resource, running the script only once for the resource
	// being managed, with the resource information piped into the metric gathering script
	PerResourceRunMode = "per-resource"
)

const (
	defaultInterval     = 15000
	defaultNamespace    = "default"
	defaultMinReplicas  = 1
	defaultMaxReplicas  = 10
	defaultStartTime    = 1
	defaultRunMode      = PerPodRunMode
	defaultLogVerbosity = 0
)

const (
	defaultAPIEnabled = true
	defaultUseHTTPS   = false
	defaultHost       = "0.0.0.0"
	defaultPort       = 5000
	defaultCertFile   = ""
	defaultKeyFile    = ""
)

const jsonStructTag = "json"

// Config is the configuration options for the CPA
type Config struct {
	ScaleTargetRef *autoscaling.CrossVersionObjectReference `json:"scaleTargetRef"`
	PreMetric      *Method                                  `json:"preMetric"`
	PostMetric     *Method                                  `json:"postMetric"`
	PreEvaluate    *Method                                  `json:"preEvaluate"`
	PostEvaluate   *Method                                  `json:"postEvaluate"`
	PreScale       *Method                                  `json:"preScale"`
	PostScale      *Method                                  `json:"postScale"`
	Evaluate       *Method                                  `json:"evaluate"`
	Metric         *Method                                  `json:"metric"`
	Interval       int                                      `json:"interval"`
	Namespace      string                                   `json:"namespace"`
	MinReplicas    int32                                    `json:"minReplicas"`
	MaxReplicas    int32                                    `json:"maxReplicas"`
	RunMode        string                                   `json:"runMode"`
	StartTime      int64                                    `json:"startTime"`
	LogVerbosity   int32                                    `json:"logVerbosity"`
	APIConfig      *APIConfig                               `json:"apiConfig"`
}

// APIConfig is configuration options specifically for the API exposed by the CPA
type APIConfig struct {
	Enabled  bool   `json:"enabled"`
	UseHTTPS bool   `json:"useHTTPS"`
	Port     int    `json:"port"`
	Host     string `json:"host"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}

// Method describes a method for passing data/triggerering logic, such as through a shell
// command
type Method struct {
	Type    string `json:"type"`
	Timeout int    `json:"timeout"`
	Shell   *Shell `json:"shell"`
}

// Shell describes configuration options for a shell command method
type Shell struct {
	Command    []string `json:"command"`
	Entrypoint string   `json:"entrypoint"`
}

// LoadConfig loads in the default configuration, then overrides it from the config file,
// then any env vars set.
func LoadConfig(configFileData []byte, envVars map[string]string) (*Config, error) {
	config := newDefaultConfig()
	err := loadFromBytes(configFileData, config)
	if err != nil {
		return nil, err
	}
	err = loadFromEnv(config, envVars)
	if err != nil {
		return nil, err
	}
	// Check API defaults

	return config, nil
}

func loadFromBytes(data []byte, config *Config) error {
	// If no bytes file data provided, skip trying to parse it
	if data == nil {
		return nil
	}
	err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 10).Decode(config)
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

		// Extract JSON tag from the type, e.g `json:"example"` would return example
		tag := fieldType.Tag.Get(jsonStructTag)

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

		// If the type is not one of the primitives above, it must be in JSON or YAML form, so try to parse
		// it and set the value from the unmarshalled JSON or YAML value
		fieldRef := reflect.New(fieldType.Type)
		err := yaml.NewYAMLOrJSONDecoder(strings.NewReader(value), 10).Decode(fieldRef.Interface())
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
		Interval:    defaultInterval,
		Namespace:   defaultNamespace,
		MinReplicas: defaultMinReplicas,
		MaxReplicas: defaultMaxReplicas,
		StartTime:   defaultStartTime,
		RunMode:     defaultRunMode,
		APIConfig: &APIConfig{
			Enabled:  defaultAPIEnabled,
			UseHTTPS: defaultUseHTTPS,
			Port:     defaultPort,
			Host:     defaultHost,
			CertFile: defaultCertFile,
			KeyFile:  defaultKeyFile,
		},
	}
}
