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

// Package metric provides functionality for managing gathering metrics,
// calling external metric gathering logic through shell commands with
// relevant data piped to them.
package metric

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
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
	RunType  string        `json:"runType"`
	Metrics  []*Metric     `json:"metrics"`
	Resource metav1.Object `json:"resource"`
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
	glog.V(3).Infoln("Gathering metrics in per-resource mode")

	// Convert the Resource description to JSON
	resourceJSON, err := json.Marshal(resource)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	if m.Config.PreMetric != nil {
		glog.V(3).Infoln("Attempting to run pre-metric hook")
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PreMetric, string(resourceJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Pre-metric hook response: %+v", hookResult)
	}

	glog.V(3).Infoln("Attempting to run metric gathering logic")
	gathered, err := m.Execute.ExecuteWithValue(m.Config.Metric, string(resourceJSON))
	if err != nil {
		return nil, err
	}
	metrics := []*Metric{
		&Metric{
			Resource: resource.GetName(),
			Value:    gathered,
		},
	}
	glog.V(3).Infof("Metrics gathered: %+v", gathered)

	if m.Config.PostMetric != nil {
		glog.V(3).Infoln("Attempting to run post-metric hook")
		postMetric := struct {
			Resource metav1.Object `json:"resource"`
			Metrics  []*Metric     `json:"metrics"`
		}{
			Resource: resource,
			Metrics:  metrics,
		}
		// Convert post metrics into JSON
		postMetricJSON, err := json.Marshal(postMetric)
		if err != nil {
			// Should not occur, panic
			panic(err)
		}
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PostMetric, string(postMetricJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Post-metric hook response: %+v", hookResult)
	}

	return &ResourceMetrics{
		Resource: resource,
		Metrics:  metrics,
	}, nil
}

func (m *Gatherer) getMetricsForPods(resource metav1.Object) (*ResourceMetrics, error) {
	glog.V(3).Infoln("Gathering metrics in per-pod mode")

	glog.V(3).Infoln("Attempting to get pod selector from managed resource")
	labels, err := m.getPodSelectorForResource(resource)
	if err != nil {
		return nil, err
	}
	glog.V(3).Infof("Label selector retrieved: %+v", labels)

	glog.V(3).Infoln("Attempting to get pods being managed")
	pods, err := m.Clientset.CoreV1().Pods(m.Config.Namespace).List(metav1.ListOptions{LabelSelector: labels})
	if err != nil {
		return nil, err
	}
	glog.V(3).Infof("Pods retrieved: %+v", pods)

	// Convert the Pods descriptions to JSON
	podsJSON, err := json.Marshal(pods.Items)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	if m.Config.PreMetric != nil {
		glog.V(3).Infoln("Attempting to run pre-metric hook")
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PreMetric, string(podsJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Pre-metric hook response: %+v", hookResult)
	}

	glog.V(3).Infoln("Attempting to gather metrics for each pod")
	var metrics []*Metric
	for _, pod := range pods.Items {
		// Convert the Pod description to JSON
		podJSON, err := json.Marshal(pod)
		if err != nil {
			// Should not occur, panic
			panic(err)
		}

		glog.V(3).Infof("Running metric gathering for pod: %s", pod.Name)
		gathered, err := m.Execute.ExecuteWithValue(m.Config.Metric, string(podJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Metric gathered: %+v", gathered)

		// Add metric to metrics array
		metrics = append(metrics, &Metric{
			Resource: pod.GetName(),
			Value:    gathered,
		})
	}
	glog.V(3).Infoln("All metrics gathered for each pod successfully")

	if m.Config.PostMetric != nil {
		glog.V(3).Infoln("Attempting to run post-metric hook")
		postMetric := struct {
			Resource metav1.Object `json:"resource"`
			Metrics  []*Metric     `json:"metrics"`
		}{
			Resource: resource,
			Metrics:  metrics,
		}
		// Convert post metrics into JSON
		postMetricJSON, err := json.Marshal(postMetric)
		if err != nil {
			// Should not occur, panic
			panic(err)
		}
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PostMetric, string(postMetricJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Post-metric hook response: %+v", hookResult)
	}

	return &ResourceMetrics{
		Resource: resource,
		Metrics:  metrics,
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
