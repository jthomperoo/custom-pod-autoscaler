/*
Copyright 2022 The Custom Pod Autoscaler Authors.

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

package fake

import (
	"time"

	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/podmetrics"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// MetricClient (fake) provides a way to insert functionality into a MetricClient
type MetricClient struct {
	GetResourceMetricReactor func(resource corev1.ResourceName, namespace string, selector labels.Selector) (podmetrics.MetricsInfo, time.Time, error)
	GetRawMetricReactor      func(metricName string, namespace string, selector labels.Selector, metricSelector labels.Selector) (podmetrics.MetricsInfo, time.Time, error)
	GetObjectMetricReactor   func(metricName string, namespace string, objectRef *autoscalingv2.CrossVersionObjectReference, metricSelector labels.Selector) (int64, time.Time, error)
	GetExternalMetricReactor func(metricName string, namespace string, selector labels.Selector) ([]int64, time.Time, error)
}

// GetResourceMetric calls the fake MetricClient function
func (f *MetricClient) GetResourceMetric(resource corev1.ResourceName, namespace string, selector labels.Selector) (podmetrics.MetricsInfo, time.Time, error) {
	return f.GetResourceMetricReactor(resource, namespace, selector)
}

// GetRawMetric calls the fake MetricClient function
func (f *MetricClient) GetRawMetric(metricName string, namespace string, selector labels.Selector, metricSelector labels.Selector) (podmetrics.MetricsInfo, time.Time, error) {
	return f.GetRawMetricReactor(metricName, namespace, selector, metricSelector)
}

// GetObjectMetric calls the fake MetricClient function
func (f *MetricClient) GetObjectMetric(metricName string, namespace string, objectRef *autoscalingv2.CrossVersionObjectReference, metricSelector labels.Selector) (int64, time.Time, error) {
	return f.GetObjectMetricReactor(metricName, namespace, objectRef, metricSelector)
}

// GetExternalMetric calls the fake MetricClient function
func (f *MetricClient) GetExternalMetric(metricName string, namespace string, selector labels.Selector) ([]int64, time.Time, error) {
	return f.GetExternalMetricReactor(metricName, namespace, selector)
}
