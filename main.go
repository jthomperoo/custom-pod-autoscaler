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
package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/jthomperoo/custom-pod-autoscaler/api"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/scaler"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	configEnvName          = "config_path"
	evaluateEnvName        = "evaluate"
	metricEnvName          = "metric"
	intervalEnvName        = "interval"
	selectorEnvName        = "selector"
	hostEnvName            = "host"
	portEnvName            = "port"
	metricTimeoutEnvName   = "metric_timeout"
	evaluateTimeoutEnvName = "evaluate_timeout"
	namespaceEnvName       = "namespace"
)

const defaultConfig = "/config.yaml"

func main() {
	// Create the in-cluster config
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Panicf(err.Error())
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Panicf(err.Error())
	}

	// Read in environment variables
	configPath, exists := os.LookupEnv(configEnvName)
	if !exists {
		configPath = defaultConfig
	}
	configEnvs := readEnvVars()

	// Read in config file
	configFileData, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Panicf(err.Error())
	}

	// Load CPA config
	config, err := config.LoadConfig(configFileData, configEnvs)
	if err != nil {
		log.Panicf(err.Error())
	}

	// Set up client for managing deployments
	deploymentsClient := clientset.AppsV1().Deployments(config.Namespace)

	// Start scaler
	scaler.ConfigureScaler(clientset, deploymentsClient, config)

	// Start API
	api.ConfigureAPI(clientset, deploymentsClient, config)
}

// readEnvVars loads in all relevant environment variables if they exist,
// putting them in a key-value map
func readEnvVars() map[string]string {
	configEnvsNames := []string{
		evaluateEnvName,
		metricEnvName,
		intervalEnvName,
		selectorEnvName,
		hostEnvName,
		hostEnvName,
		portEnvName,
		namespaceEnvName,
		metricTimeoutEnvName,
		evaluateTimeoutEnvName,
	}
	configEnvs := map[string]string{}
	for _, envName := range configEnvsNames {
		value, exists := os.LookupEnv(envName)
		if exists {
			configEnvs[envName] = value
		}
	}
	return configEnvs
}
