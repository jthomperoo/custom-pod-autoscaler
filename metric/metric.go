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

package metric

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type executeWithPiper interface {
	ExecuteWithPipe(command string, value string, timeout int) (*bytes.Buffer, error)
}

// ResourceMetrics represents a resource's metrics, including each pod's metrics
type ResourceMetrics struct {
	DeploymentName string             `json:"deployment"`
	Metrics        []*Metric          `json:"metrics"`
	Deployment     *appsv1.Deployment `json:"-"` // hide
}

// Metric is the result of the custom metric calculation, containing information on the
// relevant pod and the metric value
type Metric struct {
	Pod   string `json:"pod,omitempty"`
	Value string `json:"value,omitempty"`
}

// Gatherer handles triggering the metric gathering logic to gather metrics for a resource
type Gatherer struct {
	Clientset kubernetes.Interface
	Config    *config.Config
	Executer  executeWithPiper
}

// GetMetrics gathers metrics for the deployments supplied
func (m *Gatherer) GetMetrics(deployment *appsv1.Deployment) (*ResourceMetrics, error) {
	// Get Deployment pods
	labels := deployment.GetLabels()
	pods, err := m.Clientset.CoreV1().Pods(m.Config.Namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", labels["app"])})
	if err != nil {
		return nil, err
	}

	// Gather metrics for each pod
	var metrics []*Metric
	for _, pod := range pods.Items {
		// Convert the Pod description to JSON
		podJSON, err := json.Marshal(pod)
		if err != nil {
			// Should not occur, panic
			log.Panic(err)
		}

		// Execute the Metric command with the Pod JSON
		outb, err := m.Executer.ExecuteWithPipe(m.Config.Metric, string(podJSON), m.Config.MetricTimeout)
		if err != nil {
			log.Println(outb.String())
			return nil, err
		}

		// Add metric to metrics array
		metrics = append(metrics, &Metric{
			Pod:   pod.GetName(),
			Value: outb.String(),
		})
	}

	return &ResourceMetrics{
		DeploymentName: deployment.GetName(),
		Deployment:     deployment,
		Metrics:        metrics,
	}, nil
}
