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

Modifications Copyright 2021 The Custom Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.

Modified to split up evaluations and metric gathering to work with the
Custom Pod Autoscaler framework.
Original source:
https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/podautoscaler/horizontal.go
https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/podautoscaler/replica_calculator.go
*/

package external

import (
	"fmt"
	"time"

	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/podutil"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/value"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/labels"
	metricsclient "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

// Gatherer (External) allows retrieval of external metrics.
type Gatherer interface {
	GetMetric(metricName, namespace string, metricSelector *metav1.LabelSelector, podSelector labels.Selector) (*Metric, error)
	GetPerPodMetric(metricName, namespace string, metricSelector *metav1.LabelSelector) (*Metric, error)
}

// Metric (External) is a global metric that is not associated
// with any Kubernetes object. It allows autoscaling based on information
// coming from components running outside of cluster
// (for example length of queue in cloud messaging service, or
// QPS from loadbalancer running outside of cluster).
type Metric struct {
	Current       value.MetricValue
	ReadyPodCount *int64
	Timestamp     time.Time
}

// Gather (External)  provides functionality for retrieving metrics for external metric specs.
type Gather struct {
	MetricsClient   metricsclient.MetricsClient
	PodReadyCounter podutil.PodReadyCounter
}

// GetMetric retrieves an external metric
func (c *Gather) GetMetric(metricName, namespace string, metricSelector *metav1.LabelSelector, podSelector labels.Selector) (*Metric, error) {
	// Convert selector to expected type
	metricLabelSelector, err := metav1.LabelSelectorAsSelector(metricSelector)
	if err != nil {
		return nil, err
	}

	// Get metrics
	metrics, timestamp, err := c.MetricsClient.GetExternalMetric(metricName, namespace, metricLabelSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to get external metric %s/%s/%+v: %s", namespace, metricName, metricSelector, err)
	}
	utilization := int64(0)
	for _, val := range metrics {
		utilization = utilization + val
	}

	// Calculate number of ready pods
	readyPodCount, err := c.PodReadyCounter.GetReadyPodsCount(namespace, podSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to calculate ready pods: %s", err)
	}

	return &Metric{
		Current: value.MetricValue{
			Value: &utilization,
		},
		ReadyPodCount: &readyPodCount,
		Timestamp:     timestamp,
	}, nil
}

// GetPerPodMetric retrieves an external per pod metric
func (c *Gather) GetPerPodMetric(metricName, namespace string, metricSelector *metav1.LabelSelector) (*Metric, error) {
	// Convert selector to expected type
	metricLabelSelector, err := metav1.LabelSelectorAsSelector(metricSelector)
	if err != nil {
		return nil, err
	}

	// Get metrics
	metrics, timestamp, err := c.MetricsClient.GetExternalMetric(metricName, namespace, metricLabelSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to get external metric %s/%s/%+v: %s", namespace, metricName, metricSelector, err)
	}

	// Calculate utilization total for pods
	utilization := int64(0)
	for _, val := range metrics {
		utilization = utilization + val
	}

	return &Metric{
		Current: value.MetricValue{
			AverageValue: &utilization,
		},
		Timestamp: timestamp,
	}, nil
}
