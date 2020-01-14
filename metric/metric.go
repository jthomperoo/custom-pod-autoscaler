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

// Package metric provides functionality for managing gathering metrics,
// calling external metric gathering logic through shell commands with
// relevant data piped to them.
package metric

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/execute"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// GetMetricer provides methods for retrieving metrics
type GetMetricer interface {
	GetMetrics(resource metav1.Object) (*ResourceMetrics, error)
}

// ResourceMetrics represents a resource's metrics, including each resource's metrics
type ResourceMetrics struct {
	ResourceName string        `json:"resource"`
	RunType      string        `json:"run_type"`
	Metrics      []*Metric     `json:"metrics"`
	Resource     metav1.Object `json:"-"` // hide
}

// Metric is the result of the custom metric calculation, containing information on the
// relevant resource and the metric value
type Metric struct {
	Resource string `json:"resource,omitempty"`
	Value    string `json:"value,omitempty"`
}

// Gatherer handles triggering the metric gathering logic to gather metrics for a resource
type Gatherer struct {
	Clientset kubernetes.Interface
	Config    *config.Config
	Execute   execute.Executer
}

// GetMetrics gathers metrics for the resource supplied
func (m *Gatherer) GetMetrics(resource metav1.Object) (*ResourceMetrics, error) {
	switch m.Config.RunMode {
	case config.PerPodRunMode:
		return m.getMetricsForPods(resource)
	case config.PerResourceRunMode:
		return m.getMetricsForResource(resource)
	default:
		return nil, fmt.Errorf("Unknown run mode: %s", m.Config.RunMode)
	}
}

func (m *Gatherer) getMetricsForResource(resource metav1.Object) (*ResourceMetrics, error) {
	// Convert the Resource description to JSON
	resourceJSON, err := json.Marshal(resource)
	if err != nil {
		// Should not occur, panic
		log.Panic(err)
	}

	// Execute with the value
	gathered, err := m.Execute.ExecuteWithValue(m.Config.Metric, string(resourceJSON))
	if err != nil {
		return nil, err
	}

	return &ResourceMetrics{
		ResourceName: resource.GetName(),
		Resource:     resource,
		Metrics: []*Metric{
			&Metric{
				Resource: resource.GetName(),
				Value:    gathered,
			},
		},
	}, nil
}

func (m *Gatherer) getMetricsForPods(resource metav1.Object) (*ResourceMetrics, error) {
	// Get Resource pod selector
	labels, err := m.getPodSelectorForResource(resource)
	if err != nil {
		return nil, err
	}

	// Get Resource pods
	pods, err := m.Clientset.CoreV1().Pods(m.Config.Namespace).List(metav1.ListOptions{LabelSelector: labels})
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

		// Execute with the value
		gathered, err := m.Execute.ExecuteWithValue(m.Config.Metric, string(podJSON))
		if err != nil {
			return nil, err
		}

		// Add metric to metrics array
		metrics = append(metrics, &Metric{
			Resource: pod.GetName(),
			Value:    gathered,
		})
	}
	return &ResourceMetrics{
		ResourceName: resource.GetName(),
		Resource:     resource,
		Metrics:      metrics,
	}, nil
}

func (m *Gatherer) getPodSelectorForResource(resource metav1.Object) (string, error) {
	switch v := resource.(type) {
	case *appsv1.Deployment:
		selector, err := metav1.LabelSelectorAsMap(v.Spec.Selector)
		if err != nil {
			return "", err
		}
		return labels.SelectorFromSet(selector).String(), nil
	case *appsv1.ReplicaSet:
		selector, err := metav1.LabelSelectorAsMap(v.Spec.Selector)
		if err != nil {
			return "", err
		}
		return labels.SelectorFromSet(selector).String(), nil
	case *appsv1.StatefulSet:
		selector, err := metav1.LabelSelectorAsMap(v.Spec.Selector)
		if err != nil {
			return "", err
		}
		return labels.SelectorFromSet(selector).String(), nil
	case *corev1.ReplicationController:
		return labels.SelectorFromSet(v.Spec.Selector).String(), nil
	default:
		return "", fmt.Errorf("Unsupported resource of type %T", v)
	}
}
