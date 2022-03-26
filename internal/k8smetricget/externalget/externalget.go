/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Modifications Copyright 2022 The Custom Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.

Modified to split up evaluations and metric gathering to work with the
Custom Pod Autoscaler framework.
Original source:
https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/podautoscaler/horizontal.go
https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/podautoscaler/replica_calculator.go
*/

package externalget

import (
	"fmt"

	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/podutil"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/external"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/value"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metricsclient "github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/metrics"
	"k8s.io/apimachinery/pkg/labels"
)

// Gatherer (External) allows retrieval of external metrics.
type Gatherer interface {
	GetMetric(metricName, namespace string, metricSelector *metav1.LabelSelector, podSelector labels.Selector) (*external.Metric, error)
	GetPerPodMetric(metricName, namespace string, metricSelector *metav1.LabelSelector) (*external.Metric, error)
}

// Gather (External)  provides functionality for retrieving metrics for external metric specs.
type Gather struct {
	MetricsClient   metricsclient.Client
	PodReadyCounter podutil.PodReadyCounter
}

// GetMetric retrieves an external metric
func (c *Gather) GetMetric(metricName, namespace string, metricSelector *metav1.LabelSelector, podSelector labels.Selector) (*external.Metric, error) {
	// Convert selector to expected type
	metricLabelSelector, err := metav1.LabelSelectorAsSelector(metricSelector)
	if err != nil {
		return nil, err
	}

	// Get metrics
	metrics, timestamp, err := c.MetricsClient.GetExternalMetric(metricName, namespace, metricLabelSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to get external metric %s/%s/%+v: %w", namespace, metricName, metricSelector, err)
	}
	utilization := int64(0)
	for _, val := range metrics {
		utilization = utilization + val
	}

	// Calculate number of ready pods
	readyPodCount, err := c.PodReadyCounter.GetReadyPodsCount(namespace, podSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to calculate ready pods: %w", err)
	}

	return &external.Metric{
		Current: value.MetricValue{
			Value: &utilization,
		},
		ReadyPodCount: &readyPodCount,
		Timestamp:     timestamp,
	}, nil
}

// GetPerPodMetric retrieves an external per pod metric
func (c *Gather) GetPerPodMetric(metricName, namespace string, metricSelector *metav1.LabelSelector) (*external.Metric, error) {
	// Convert selector to expected type
	metricLabelSelector, err := metav1.LabelSelectorAsSelector(metricSelector)
	if err != nil {
		return nil, err
	}

	// Get metrics
	metrics, timestamp, err := c.MetricsClient.GetExternalMetric(metricName, namespace, metricLabelSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to get external metric %s/%s/%+v: %w", namespace, metricName, metricSelector, err)
	}

	// Calculate utilization total for pods
	utilization := int64(0)
	for _, val := range metrics {
		utilization = utilization + val
	}

	return &external.Metric{
		Current: value.MetricValue{
			AverageValue: &utilization,
		},
		Timestamp: timestamp,
	}, nil
}
