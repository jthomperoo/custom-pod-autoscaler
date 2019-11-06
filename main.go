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
	"os/exec"

	"github.com/jthomperoo/custom-pod-autoscaler/api"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/scaler"
	"github.com/jthomperoo/custom-pod-autoscaler/shell"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	configEnvName          = "config_path"
	evaluateEnvName        = "evaluate"
	metricEnvName          = "metric"
	intervalEnvName        = "interval"
	hostEnvName            = "host"
	portEnvName            = "port"
	metricTimeoutEnvName   = "metricTimeout"
	evaluateTimeoutEnvName = "evaluateTimeout"
	namespaceEnvName       = "namespace"
	scaleTargetRefEnvName  = "scaleTargetRef"
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

	// Set up shell execution context
	executer := shell.NewCommandExecuteWithPipe(exec.Command)

	// Start scaler
	scaler.ConfigureScaler(clientset, deploymentsClient, config, executer)

	// Start API
	api.ConfigureAPI(clientset, deploymentsClient, config, executer)
}

// readEnvVars loads in all relevant environment variables if they exist,
// putting them in a key-value map
func readEnvVars() map[string]string {
	configEnvsNames := []string{
		evaluateEnvName,
		metricEnvName,
		intervalEnvName,
		hostEnvName,
		hostEnvName,
		portEnvName,
		namespaceEnvName,
		metricTimeoutEnvName,
		evaluateTimeoutEnvName,
		scaleTargetRefEnvName,
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
