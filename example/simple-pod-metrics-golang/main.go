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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	corev1 "k8s.io/api/core/v1"
)

// MetricValue is a representation of the metric retrieved from from the 'flask-metric' application
type MetricValue struct {
	Available int `json:"available"`
	Value     int `json:"value"`
	Min       int `json:"min"`
	Max       int `json:"max"`
}

func main() {
	// Read in stdin
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Determine if gathering metrics or evaluating based on flag
	modePtr := flag.String("mode", "no_mode", "command mode, either metric or evaluate")
	flag.Parse()

	switch *modePtr {
	case "metric":
		getMetrics(stdin)
	case "evaluate":
		getEvaluation(stdin)
	default:
		log.Fatalf("Unknown command mode: %s", *modePtr)
		os.Exit(1)
	}
}

func getMetrics(stdin []byte) {
	// Attempt to unmarshal stdin resource description into a Pod definition
	var pod corev1.Pod
	err := json.Unmarshal(stdin, &pod)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Make a HTTP request to the pod's '/metric' endpoint
	client := http.Client{}
	resp, err := client.Get(fmt.Sprintf("http://%s:5000/metric", pod.Status.PodIP))
	if err != nil {
		log.Fatalf("Error occurred retrieving metrics: %s", err)
	}

	// Read HTTP request response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error occurred reading result body: %s", err)
	}

	// If not 200 response, error, otherwise print the response to stdout
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error occurred, non 200 response code, code %d: %s", resp.StatusCode, string(body))
	}

	fmt.Print(string(body))
}

func getEvaluation(stdin []byte) {
	var resourceMetrics metric.ResourceMetrics
	err := json.Unmarshal(stdin, &resourceMetrics)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Count total available
	totalAvailable := 0
	for _, metric := range resourceMetrics.Metrics {
		var metricValue MetricValue
		err := json.Unmarshal([]byte(metric.Value), &metricValue)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		totalAvailable += metricValue.Available
	}

	// Get current replica count
	targetReplicaCount := len(resourceMetrics.Metrics)

	// Decrease target rpelicas if more than 5 available
	if totalAvailable > 5 {
		targetReplicaCount--
	}

	// Increase target replicas if none available
	if totalAvailable <= 0 {
		targetReplicaCount++
	}

	// Build JSON response
	evaluation := evaluate.Evaluation{
		TargetReplicas: int32(targetReplicaCount),
	}

	// Output JSON to stdout
	output, err := json.Marshal(evaluation)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Print(string(output))
}
