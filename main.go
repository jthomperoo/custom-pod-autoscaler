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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/jthomperoo/custom-pod-autoscaler/api"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/scaler"
	"github.com/jthomperoo/custom-pod-autoscaler/shell"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	configEnvName          = "configPath"
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
	// Read in environment variables
	configPath, exists := os.LookupEnv(configEnvName)
	if !exists {
		configPath = defaultConfig
	}
	configEnvs := readEnvVars()

	// Read in config file
	configFileData, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Panic(err)
	}

	// Load Custom Pod Autoscaler config
	config, err := config.LoadConfig(configFileData, configEnvs)
	if err != nil {
		log.Panic(err)
	}

	// Create the in-cluster Kubernetes config
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Panic(err)
	}

	// Create the Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Panic(err)
	}

	// Set up client for managing Kubernetes deployments
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

	// Set up time ticker with configured interval
	ticker := time.NewTicker(time.Duration(config.Interval) * time.Millisecond)
	// Set up shutdown channel, which will listen for UNIX shutdown commands
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Set up scaler
	autoscaler := &scaler.Scaler{
		Clientset:         clientset,
		DeploymentsClient: deploymentsClient,
		Config:            config,
		Ticker:            ticker,
		Shutdown:          shutdown,
		GetMetricer:       metricGatherer,
		GetEvaluationer:   evaluator,
	}

	// Run the scaler in a goroutine, triggered by the ticker
	go func() {
		for {
			select {
			case <-shutdown:
				ticker.Stop()
				return
			case <-ticker.C:
				err := autoscaler.Scale()
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	// Set up API
	api := &api.API{
		Router:          chi.NewRouter(),
		Config:          config,
		Clientset:       clientset,
		GetMetricer:     metricGatherer,
		GetEvaluationer: evaluator,
	}
	api.Routes()
	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%d", api.Config.Host, api.Config.Port),
		Handler: api.Router,
	}

	// Start API
	srv.ListenAndServe()
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
