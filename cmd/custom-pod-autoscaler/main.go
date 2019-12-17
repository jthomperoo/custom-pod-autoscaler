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

// Custom Pod Autoscaler is the core program that runs inside a Custom Pod Autoscaler Image.
// The program handles interactions with the Kubernetes API, manages triggering Custom Pod
// Autoscaler User Logic through shell commands, exposes a simple HTTP REST API for viewing
// metrics and evaluations, and handles parsing user configuration to specify polling intervals,
// Kubernetes namespaces, command timeouts etc.
// The Custom Pod Autoscaler must be run inside a Kubernetes cluster.
package main

import (
	"context"
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
	"github.com/jthomperoo/custom-pod-autoscaler/execute"
	"github.com/jthomperoo/custom-pod-autoscaler/execute/shell"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/scaler"
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
	runModeEnvName         = "runMode"
	startTimeEnvName       = "startTime"
	minReplicasEnvName     = "minReplicas"
	maxReplicasEnvName     = "maxReplicas"
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
	shellExec := &shell.Execute{
		Command: exec.Command,
	}

	// Set up metric gathering
	metricGatherer := &metric.Gatherer{
		Clientset: clientset,
		Config:    config,
		Execute: &execute.CombinedExecute{
			Executers: []execute.Executer{
				shellExec,
			},
		},
	}

	// Set up evaluator
	evaluator := &evaluate.Evaluator{
		Config: config,
		Execute: &execute.CombinedExecute{
			Executers: []execute.Executer{
				shellExec,
			},
		},
	}

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

	delayTime := config.StartTime - (time.Now().UTC().UnixNano() / int64(time.Millisecond) % config.StartTime)
	delayStartTimer := time.NewTimer(time.Duration(delayTime) * time.Millisecond)

	log.Printf("Waiting %d milliseconds before starting autoscaler\n", delayTime)

	go func() {
		// Wait for delay to start at expected time
		<-delayStartTimer.C
		log.Println("Starting autoscaler")
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
			GetMetricer:       metricGatherer,
			GetEvaluationer:   evaluator,
		}

		// Run the scaler in a goroutine, triggered by the ticker
		// listen for shutdown requests, once recieved shut down the autoscaler
		// and the API
		go func() {
			for {
				select {
				case <-shutdown:
					log.Println("Shutting down...")
					// Stop API
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					srv.Shutdown(ctx)
					// Stop autoscaler
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
	}()

	log.Println("Starting API")
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
		runModeEnvName,
		minReplicasEnvName,
		maxReplicasEnvName,
		startTimeEnvName,
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
