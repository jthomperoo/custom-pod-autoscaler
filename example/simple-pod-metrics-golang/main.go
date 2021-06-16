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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/metric"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

// EvaluateSpec represents the information fed to the evaluator
type EvaluateSpec struct {
	Metrics              []*metric.ResourceMetric  `json:"metrics"`
	UnstructuredResource unstructured.Unstructured `json:"resource"`
	Resource             metav1.Object             `json:"-"`
	RunType              string                    `json:"runType"`
}

// MetricSpec represents the information fed to the metric gatherer
type MetricSpec struct {
	Pod     corev1.Pod `json:"resource"`
	RunType string     `json:"runType"`
}

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
		log.Println("gathering metrics")
		getMetrics(stdin)
	case "evaluate":
		log.Println("evaluating metrics")
		getEvaluation(stdin)
	default:
		log.Fatalf("Unknown command mode: %s", *modePtr)
		os.Exit(1)
	}
}

func getMetrics(stdin []byte) {
	log.Println("parsing spec")
	var spec MetricSpec
	err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(stdin), 10).Decode(&spec)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Println("sending HTTP request to gather pod metrics")
	// Make a HTTP request to the pod's '/metric' endpoint
	client := http.Client{}
	resp, err := client.Get(fmt.Sprintf("http://%s:5000/metric", spec.Pod.Status.PodIP))
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

	log.Printf("recieved pod metrics: %s\n", string(body))

	fmt.Print(string(body))
}

func getEvaluation(stdin []byte) {
	log.Println("parsing spec")
	var spec EvaluateSpec
	err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(stdin), 10).Decode(&spec)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Create object from version and kind of piped value
	resourceGVK := spec.UnstructuredResource.GroupVersionKind()
	resourceRuntime, err := scheme.Scheme.New(resourceGVK)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Parse the unstructured k8s resource into the object created, then convert to generic metav1.Object
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(spec.UnstructuredResource.Object, resourceRuntime)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	spec.Resource = resourceRuntime.(metav1.Object)

	log.Println("calculating total available")
	// Count total available
	totalAvailable := 0
	for _, metric := range spec.Metrics {
		var metricValue MetricValue
		err := json.Unmarshal([]byte(metric.Value), &metricValue)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		totalAvailable += metricValue.Available
	}

	// Get current replica count
	targetReplicaCount := len(spec.Metrics)

	log.Printf("initial replica count: %d\n", targetReplicaCount)

	// Decrease target rpelicas if more than 5 available
	if totalAvailable > 5 {
		log.Printf("decreasing target replica count since >5 available to: %d\n", targetReplicaCount)
		targetReplicaCount--
	}

	// Increase target replicas if none available
	if totalAvailable <= 0 {
		log.Printf("increasing target replica count since 0 available to: %d\n", targetReplicaCount)
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
