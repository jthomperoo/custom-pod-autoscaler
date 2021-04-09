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

package fake

import (
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/k8smetric"
	"github.com/jthomperoo/custom-pod-autoscaler/k8smetric/external"
	"github.com/jthomperoo/custom-pod-autoscaler/k8smetric/object"
	"github.com/jthomperoo/custom-pod-autoscaler/k8smetric/pods"
	"github.com/jthomperoo/custom-pod-autoscaler/k8smetric/resource"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// Gather (fake) provides a way to insert functionality into a Gatherer
type Gather struct {
	GetMetricsReactor func(resource metav1.Object, specs []config.K8sMetricSpec, namespace string) ([]*k8smetric.Metric, error)
}

// GetMetrics calls the fake Gather function
func (f *Gather) GetMetrics(resource metav1.Object, specs []config.K8sMetricSpec, namespace string) ([]*k8smetric.Metric, error) {
	return f.GetMetricsReactor(resource, specs, namespace)
}

// ExternalGatherer (fake) provides a way to insert functionality into an ExternalGatherer
type ExternalGatherer struct {
	GetMetricReactor       func(metricName, namespace string, metricSelector *metav1.LabelSelector, podSelector labels.Selector) (*external.Metric, error)
	GetPerPodMetricReactor func(metricName, namespace string, metricSelector *metav1.LabelSelector) (*external.Metric, error)
}

// GetMetric calls the fake ExternalGatherer function
func (f *ExternalGatherer) GetMetric(metricName, namespace string, metricSelector *metav1.LabelSelector, podSelector labels.Selector) (*external.Metric, error) {
	return f.GetMetricReactor(metricName, namespace, metricSelector, podSelector)
}

// GetPerPodMetric calls the fake ExternalGatherer function
func (f *ExternalGatherer) GetPerPodMetric(metricName, namespace string, metricSelector *metav1.LabelSelector) (*external.Metric, error) {
	return f.GetPerPodMetricReactor(metricName, namespace, metricSelector)
}

// ObjectGatherer (fake) provides a way to insert functionality into an ObjectGatherer
type ObjectGatherer struct {
	GetMetricReactor       func(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, selector labels.Selector, metricSelector labels.Selector) (*object.Metric, error)
	GetPerPodMetricReactor func(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (*object.Metric, error)
}

// GetMetric calls the fake ObjectGatherer function
func (f *ObjectGatherer) GetMetric(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, selector labels.Selector, metricSelector labels.Selector) (*object.Metric, error) {
	return f.GetMetricReactor(metricName, namespace, objectRef, selector, metricSelector)
}

// GetPerPodMetric calls the fake ObjectGatherer function
func (f *ObjectGatherer) GetPerPodMetric(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (*object.Metric, error) {
	return f.GetPerPodMetricReactor(metricName, namespace, objectRef, metricSelector)
}

// PodsGatherer (fake) provides a way to insert functionality into an PodsGatherer
type PodsGatherer struct {
	GetMetricReactor func(metricName string, namespace string, selector labels.Selector, metricSelector labels.Selector) (*pods.Metric, error)
}

// GetMetric calls the fake PodsGatherer function
func (f *PodsGatherer) GetMetric(metricName string, namespace string, selector labels.Selector, metricSelector labels.Selector) (*pods.Metric, error) {
	return f.GetMetricReactor(metricName, namespace, selector, metricSelector)
}

// ResourceGatherer (fake) provides a way to insert functionality into an ResourceGatherer
type ResourceGatherer struct {
	GetMetricReactor    func(resource corev1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error)
	GetRawMetricReactor func(resource corev1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error)
}

// GetMetric calls the fake ResourceGatherer function
func (f *ResourceGatherer) GetMetric(resource corev1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error) {
	return f.GetMetricReactor(resource, namespace, selector)
}

// GetRawMetric calls the fake ResourceGatherer function
func (f *ResourceGatherer) GetRawMetric(resource corev1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error) {
	return f.GetRawMetricReactor(resource, namespace, selector)
}
