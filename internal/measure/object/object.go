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

package object

import (
	"fmt"
	"time"

	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/podutil"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/labels"
	metricsclient "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

// Gatherer (Object) allows retrieval of object metrics.
type Gatherer interface {
	GetMetric(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, selector labels.Selector, metricSelector labels.Selector) (*Metric, error)
	GetPerPodMetric(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (*Metric, error)
}

// Metric (Object) is a metric describing a kubernetes object
// (for example, hits-per-second on an Ingress object).
type Metric struct {
	Utilization   int64
	ReadyPodCount *int64
	Timestamp     time.Time
}

// Gather (Object) provides functionality for retrieving metrics for object metric specs.
type Gather struct {
	MetricsClient   metricsclient.MetricsClient
	PodReadyCounter podutil.PodReadyCounter
}

// GetMetric retrieves an object metric
func (c *Gather) GetMetric(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, selector labels.Selector, metricSelector labels.Selector) (*Metric, error) {
	// Get metrics
	utilization, timestamp, err := c.MetricsClient.GetObjectMetric(metricName, namespace, objectRef, metricSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to get metric %s: %v on %s %s/%s", metricName, objectRef.Kind, namespace, objectRef.Name, err)
	}

	// Calculate number of ready pods
	readyPodCount, err := c.PodReadyCounter.GetReadyPodsCount(namespace, selector)
	if err != nil {
		return nil, fmt.Errorf("unable to calculate ready pods: %s", err)
	}

	return &Metric{
		Utilization:   utilization,
		ReadyPodCount: &readyPodCount,
		Timestamp:     timestamp,
	}, nil
}

// GetPerPodMetric retrieves an object per pod metric
func (c *Gather) GetPerPodMetric(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (*Metric, error) {
	// Get metrics
	utilization, timestamp, err := c.MetricsClient.GetObjectMetric(metricName, namespace, objectRef, metricSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to get metric %s: %v on %s %s/%s", metricName, objectRef.Kind, namespace, objectRef.Name, err)
	}

	return &Metric{
		Utilization: utilization,
		Timestamp:   timestamp,
	}, nil
}
