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
	"github.com/jthomperoo/custom-pod-autoscaler/autoscaler"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
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

	// Set up shell executer
	executer := shell.ExecuteWithPipe{
		Command: exec.Command,
	}

	// Set up metric gathering
	metricGatherer := &metric.Gatherer{
		Clientset: clientset,
		Config:    config,
		Executer:  &executer,
	}

	// Set up evaluator
	evaluator := &evaluate.Evaluator{
		Config:   config,
		Executer: &executer,
	}

	// Set up autoscaler and start it
	autoscaler := autoscaler.NewAutoscaler(clientset, deploymentsClient, config, metricGatherer, evaluator)
	autoscaler.Start()

	// Start API
	api := &api.API{
		Config:            config,
		Clientset:         clientset,
		DeploymentsClient: deploymentsClient,
		GetMetricer:       metricGatherer,
		GetEvaluationer:   evaluator,
	}
	api.Start()
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
