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

package resourceget

import (
	"fmt"
	"time"

	"github.com/jthomperoo/custom-pod-autoscaler/internal/k8smetricget/podutil"
	"github.com/jthomperoo/custom-pod-autoscaler/k8smetric/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	corelisters "k8s.io/client-go/listers/core/v1"
	metricsclient "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

// Gatherer (Resource) allows retrieval of resource metrics.
type Gatherer interface {
	GetMetric(resourceName corev1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error)
	GetRawMetric(resourceName corev1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error)
}

// Gather (Resource) provides functionality for retrieving metrics for resource metric specs.
type Gather struct {
	MetricsClient                 metricsclient.MetricsClient
	PodLister                     corelisters.PodLister
	CPUInitializationPeriod       time.Duration
	DelayOfInitialReadinessStatus time.Duration
}

// GetMetric retrieves a resource metric
func (c *Gather) GetMetric(resourceName corev1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error) {
	// Get metrics
	metrics, timestamp, err := c.MetricsClient.GetResourceMetric(resourceName, namespace, selector, "")
	if err != nil {
		return nil, fmt.Errorf("unable to get metrics for resource %s: %v", resourceName, err)
	}

	// Get pods
	podList, err := c.PodLister.Pods(namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("unable to get pods while calculating replica count: %v", err)
	}

	totalPods := len(podList)
	if totalPods == 0 {
		return nil, fmt.Errorf("No pods returned by selector while calculating replica count")
	}

	// Remove missing pod metrics
	readyPodCount, ignoredPods, missingPods := podutil.GroupPods(podList, metrics, resourceName, c.CPUInitializationPeriod, c.DelayOfInitialReadinessStatus)
	podutil.RemoveMetricsForPods(metrics, ignoredPods)

	// Calculate requests - limits for pod resources
	requests, err := podutil.CalculatePodRequests(podList, resourceName)
	if err != nil {
		return nil, err
	}

	return &resource.Metric{
		PodMetricsInfo: metrics,
		Requests:       requests,
		ReadyPodCount:  int64(readyPodCount),
		IgnoredPods:    ignoredPods,
		MissingPods:    missingPods,
		TotalPods:      totalPods,
		Timestamp:      timestamp,
	}, nil
}

// GetRawMetric retrieves a a raw resource metric
func (c *Gather) GetRawMetric(resourceName corev1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error) {
	// Get metrics
	metrics, timestamp, err := c.MetricsClient.GetResourceMetric(resourceName, namespace, selector, "")
	if err != nil {
		return nil, fmt.Errorf("unable to get metrics for resource %s: %v", resourceName, err)
	}

	// Get pods
	podList, err := c.PodLister.Pods(namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("unable to get pods while calculating replica count: %v", err)
	}

	totalPods := len(podList)
	if totalPods == 0 {
		return nil, fmt.Errorf("No pods returned by selector while calculating replica count")
	}

	// Remove missing pod metrics
	readyPodCount, ignoredPods, missingPods := podutil.GroupPods(podList, metrics, resourceName, c.CPUInitializationPeriod, c.DelayOfInitialReadinessStatus)
	podutil.RemoveMetricsForPods(metrics, ignoredPods)

	return &resource.Metric{
		PodMetricsInfo: metrics,
		ReadyPodCount:  int64(readyPodCount),
		IgnoredPods:    ignoredPods,
		MissingPods:    missingPods,
		TotalPods:      totalPods,
		Timestamp:      timestamp,
	}, nil
}
