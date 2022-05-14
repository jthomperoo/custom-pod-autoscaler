/*
Copyright 2021 The Custom Pod Autoscaler Authors.

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

// Package metricget provides functionality for managing gathering metrics, calling external metric gathering logic
// through shell commands with relevant data piped to them.
package metricget

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/execute"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/metric"
	"github.com/jthomperoo/k8shorizmetrics/metrics"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// GetMetricer provides methods for retrieving metrics
type GetMetricer interface {
	GetMetrics(info metric.Info, podSelector labels.Selector) ([]*metric.ResourceMetric, error)
}

type K8sMetricGatherer interface {
	Gather(specs []autoscalingv2.MetricSpec, namespace string, podSelector labels.Selector) ([]*metrics.Metric, error)
}

// Gatherer handles triggering the metric gathering logic to gather metrics for a resource
type Gatherer struct {
	Clientset         kubernetes.Interface
	Config            *config.Config
	Execute           execute.Executer
	K8sMetricGatherer K8sMetricGatherer
}

// GetMetrics gathers metrics for the resource supplied
func (m *Gatherer) GetMetrics(info metric.Info, podSelector labels.Selector) ([]*metric.ResourceMetric, error) {
	// Query metrics server if requested
	if m.Config.KubernetesMetricSpecs != nil {
		glog.V(3).Infoln("K8s Metrics specs provided, attempting to query the K8s metrics server")
		gatheredMetrics, err := m.K8sMetricGatherer.Gather(
			convertCPAMetricSpecsToK8sMetricSpecs(m.Config.KubernetesMetricSpecs), m.Config.Namespace, podSelector)
		if err != nil {
			if m.Config.RequireKubernetesMetrics {
				return nil, fmt.Errorf("failed to get required Kubernetes metrics: %w", err)
			}
			glog.Errorf("Failed to retrieve K8s metrics, not required so continuing: %+v", err)
		} else {
			glog.V(3).Infof("Successfully retrieved K8s metrics: %+v", gatheredMetrics)
		}
		info.KubernetesMetrics = convertK8sMetricsToCPAK8sMetrics(gatheredMetrics)
		glog.V(3).Infoln("Finished querying the K8s metrics server")
	}

	switch m.Config.RunMode {
	case config.PerPodRunMode:
		return m.getMetricsForPods(info, podSelector)
	case config.PerResourceRunMode:
		return m.getMetricsForResource(info)
	default:
		return nil, fmt.Errorf("unknown run mode: %s", m.Config.RunMode)
	}
}

func (m *Gatherer) getMetricsForResource(info metric.Info) ([]*metric.ResourceMetric, error) {
	glog.V(3).Infoln("Gathering metrics in per-resource mode")

	// Convert the Resource description to JSON
	specJSON, err := json.Marshal(info)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	if m.Config.PreMetric != nil {
		glog.V(3).Infoln("Attempting to run pre-metric hook")
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PreMetric, string(specJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to run pre-metric hook: %w", err)
		}
		glog.V(3).Infof("Pre-metric hook response: %+v", hookResult)
	}

	glog.V(3).Infoln("Attempting to run metric gathering logic")
	gathered, err := m.Execute.ExecuteWithValue(m.Config.Metric, string(specJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to gather metrics: %w", err)
	}
	info.Metrics = []*metric.ResourceMetric{
		{
			Resource: info.Resource.GetName(),
			Value:    gathered,
		},
	}
	glog.V(3).Infof("Metrics gathered: %+v", gathered)

	if m.Config.PostMetric != nil {
		glog.V(3).Infoln("Attempting to run post-metric hook")
		// Convert post metrics into JSON
		postSpecJSON, err := json.Marshal(info)
		if err != nil {
			// Should not occur, panic
			panic(err)
		}
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PostMetric, string(postSpecJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to run post-metric hook: %w", err)
		}
		glog.V(3).Infof("Post-metric hook response: %+v", hookResult)
	}

	return info.Metrics, nil
}

func (m *Gatherer) getMetricsForPods(info metric.Info, podSelector labels.Selector) ([]*metric.ResourceMetric, error) {
	glog.V(3).Infoln("Gathering metrics in per-pod mode")

	glog.V(3).Infoln("Attempting to get pods being managed")
	pods, err := m.Clientset.CoreV1().Pods(m.Config.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: podSelector.String()})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods being managed: %w", err)
	}
	glog.V(3).Infof("Pods retrieved: %+v", pods)

	// Convert the metric info into JSON
	specJSON, err := json.Marshal(info)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	if m.Config.PreMetric != nil {
		glog.V(3).Infoln("Attempting to run pre-metric hook")
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PreMetric, string(specJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to run pre-metric hook: %w", err)
		}
		glog.V(3).Infof("Pre-metric hook response: %+v", hookResult)
	}

	glog.V(3).Infoln("Attempting to gather metrics for each pod")
	var metrics []*metric.ResourceMetric
	for _, pod := range pods.Items {
		// Convert the Pod description to JSON
		podSpecJSON, err := json.Marshal(metric.Info{
			Resource:          &pod,
			RunType:           info.RunType,
			KubernetesMetrics: info.KubernetesMetrics,
		})
		if err != nil {
			// Should not occur, panic
			panic(err)
		}

		glog.V(3).Infof("Running metric gathering for pod: %s", pod.Name)
		gathered, err := m.Execute.ExecuteWithValue(m.Config.Metric, string(podSpecJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to gather metrics: %w", err)
		}
		glog.V(3).Infof("Metric gathered: %+v", gathered)

		// Add metric to metrics array
		metrics = append(metrics, &metric.ResourceMetric{
			Resource: pod.GetName(),
			Value:    gathered,
		})
	}
	glog.V(3).Infoln("All metrics gathered for each pod successfully")
	info.Metrics = metrics

	if m.Config.PostMetric != nil {
		glog.V(3).Infoln("Attempting to run post-metric hook")
		// Convert post metrics into JSON
		postSpecJSON, err := json.Marshal(info)
		if err != nil {
			// Should not occur, panic
			panic(err)
		}
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PostMetric, string(postSpecJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to run post-metric hook: %w", err)
		}
		glog.V(3).Infof("Post-metric hook response: %+v", hookResult)
	}

	return metrics, nil
}

func convertK8sMetricsToCPAK8sMetrics(metrics []*metrics.Metric) []*k8smetric.Metric {
	cpaK8sMetrics := []*k8smetric.Metric{}
	for _, metric := range metrics {
		cpaK8sMetrics = append(cpaK8sMetrics, (*k8smetric.Metric)(metric))
	}
	return cpaK8sMetrics
}

func convertCPAMetricSpecsToK8sMetricSpecs(specs []config.K8sMetricSpec) []autoscalingv2.MetricSpec {
	k8sSpecs := []autoscalingv2.MetricSpec{}
	for _, spec := range specs {
		k8sSpecs = append(k8sSpecs, (autoscalingv2.MetricSpec)(spec))
	}
	return k8sSpecs
}
