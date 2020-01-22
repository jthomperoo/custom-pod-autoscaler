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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang/glog"
	v1 "github.com/jthomperoo/custom-pod-autoscaler/api/v1"
	"github.com/jthomperoo/custom-pod-autoscaler/autoscaler"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/execute"
	"github.com/jthomperoo/custom-pod-autoscaler/execute/shell"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/resourceclient"
	"github.com/jthomperoo/custom-pod-autoscaler/scale"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	k8sscale "k8s.io/client-go/scale"
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
	logVerbosityEnvName    = "logVerbosity"
)

const defaultConfig = "/config.yaml"

func init() {
	err := flag.Set("logtostderr", "true")
	if err != nil {
		log.Fatalf("Fail to set log to standard error flag: %s", err)
	}
	err = flag.Set("v", "0")
	if err != nil {
		log.Fatalf("Fail to set default log verbosity flag: %s", err)
	}
	flag.Parse()
}

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
		glog.Fatalf("Fail to read configuration file: %s", err)
	}

	// Load Custom Pod Autoscaler config
	config, err := config.LoadConfig(configFileData, configEnvs)
	if err != nil {
		glog.Fatalf("Fail to parse configuration: %s", err)
	}

	// Create the in-cluster Kubernetes config
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatalf("Fail to create in-cluster Kubernetes config: %s", err)
	}

	// Set up clientset
	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		glog.Fatalf("Fail to set up Kubernetes clientset: %s", err)
	}

	// Set up dynamic client
	dynamicClient, err := dynamic.NewForConfig(clusterConfig)
	if err != nil {
		glog.Fatalf("Fail to set up Kubernetes dynamic client: %s", err)
	}

	// Get group resources
	groupResources, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		glog.Fatalf("Fail get group resources: %s", err)
	}

	// Set logging level
	err = flag.Lookup("v").Value.Set(strconv.Itoa(int(config.LogVerbosity)))
	if err != nil {
		glog.Fatalf("Fail to set log verbosity: %s", err)
	}

	glog.V(1).Infoln("Setting up resources and clients")

	// Unstructured converter
	unstructuredConverted := runtime.DefaultUnstructuredConverter

	// Set up resource client
	resourceClient := &resourceclient.UnstructuredClient{
		Dynamic:               dynamicClient,
		UnstructuredConverter: unstructuredConverted,
	}

	scaleClient := k8sscale.New(
		clientset.RESTClient(),
		restmapper.NewDiscoveryRESTMapper(groupResources),
		dynamic.LegacyAPIPathResolverFunc,
		k8sscale.NewDiscoveryScaleKindResolver(
			clientset.Discovery(),
		),
	)

	scaler := &scale.Scale{
		Scaler: scaleClient,
	}

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

	glog.V(1).Infoln("Setting up REST API")

	// Set up API
	api := &v1.API{
		Router:          chi.NewRouter(),
		Config:          config,
		Client:          resourceClient,
		GetMetricer:     metricGatherer,
		GetEvaluationer: evaluator,
		Scaler:          scaler,
	}
	api.Routes()
	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%d", api.Config.Host, api.Config.Port),
		Handler: api.Router,
	}

	glog.V(1).Infoln("Setting up autoscaler")

	delayTime := config.StartTime - (time.Now().UTC().UnixNano() / int64(time.Millisecond) % config.StartTime)
	delayStartTimer := time.NewTimer(time.Duration(delayTime) * time.Millisecond)

	glog.V(0).Infof("Waiting %d milliseconds before starting autoscaler\n", delayTime)

	go func() {
		// Wait for delay to start at expected time
		<-delayStartTimer.C
		glog.V(0).Infoln("Starting autoscaler")
		// Set up time ticker with configured interval
		ticker := time.NewTicker(time.Duration(config.Interval) * time.Millisecond)
		// Set up shutdown channel, which will listen for UNIX shutdown commands
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		// Set up scaler
		autoscaler := &autoscaler.Scaler{
			Client:          resourceClient,
			Config:          config,
			GetMetricer:     metricGatherer,
			GetEvaluationer: evaluator,
			Scaler:          scaler,
		}

		// Run the scaler in a goroutine, triggered by the ticker
		// listen for shutdown requests, once recieved shut down the autoscaler
		// and the API
		go func() {
			for {
				select {
				case <-shutdown:
					glog.V(0).Infoln("Shutting down...")
					// Stop API
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					srv.Shutdown(ctx)
					// Stop autoscaler
					ticker.Stop()
					return
				case <-ticker.C:
					glog.V(2).Infoln("Running autoscaler")
					err := autoscaler.Scale()
					if err != nil {
						glog.Errorln(err)
					}
				}
			}
		}()
	}()

	glog.V(0).Infoln("Starting API")
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
		logVerbosityEnvName,
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
