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

// Package confload handles loading in configuration - parsing YAML and environment variable input into a
// Custom Pod Autoscaler configuration struct. Contains a set of defaults that can be overridden by provided YAML and
// env vars.
package confload

import (
	"bytes"
	"reflect"
	"strconv"
	"strings"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	defaultInterval                = 15000
	defaultNamespace               = "default"
	defaultMinReplicas             = 1
	defaultMaxReplicas             = 10
	defaultStartTime               = 1
	defaultRunMode                 = config.PerPodRunMode
	defaultLogVerbosity            = 0
	defaultDownscaleStabilization  = 0
	defaultCPUInitializationPeriod = 300
	defaultInitialReadinessDelay   = 30
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

// Load loads in the default configuration, then overrides it from the config file, then any env vars set.
func Load(configFileData []byte, envVars map[string]string) (*config.Config, error) {
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

func loadFromBytes(data []byte, config *config.Config) error {
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

func loadFromEnv(config *config.Config, envVars map[string]string) error {
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

func newDefaultConfig() *config.Config {
	return &config.Config{
		Interval:               defaultInterval,
		Namespace:              defaultNamespace,
		MinReplicas:            defaultMinReplicas,
		MaxReplicas:            defaultMaxReplicas,
		StartTime:              defaultStartTime,
		RunMode:                defaultRunMode,
		DownscaleStabilization: defaultDownscaleStabilization,
		APIConfig: &config.APIConfig{
			Enabled:  defaultAPIEnabled,
			UseHTTPS: defaultUseHTTPS,
			Port:     defaultPort,
			Host:     defaultHost,
			CertFile: defaultCertFile,
			KeyFile:  defaultKeyFile,
		},
		KubernetesMetricSpecs:    nil,
		RequireKubernetesMetrics: false,
		InitialReadinessDelay:    defaultInitialReadinessDelay,
		CPUInitializationPeriod:  defaultCPUInitializationPeriod,
	}
}
