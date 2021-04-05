package fake

import (
	"time"

	autoscaling "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	metricsclient "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

// MetricClient (fake) provides a way to insert functionality into a MetricClient
type MetricClient struct {
	GetResourceMetricReactor func(resource corev1.ResourceName, namespace string, selector labels.Selector, container string) (metricsclient.PodMetricsInfo, time.Time, error)
	GetRawMetricReactor      func(metricName string, namespace string, selector labels.Selector, metricSelector labels.Selector) (metricsclient.PodMetricsInfo, time.Time, error)
	GetObjectMetricReactor   func(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (int64, time.Time, error)
	GetExternalMetricReactor func(metricName string, namespace string, selector labels.Selector) ([]int64, time.Time, error)
}

// GetResourceMetric calls the fake MetricClient function
func (f *MetricClient) GetResourceMetric(resource corev1.ResourceName, namespace string, selector labels.Selector, container string) (metricsclient.PodMetricsInfo, time.Time, error) {
	return f.GetResourceMetricReactor(resource, namespace, selector, container)
}

// GetRawMetric calls the fake MetricClient function
func (f *MetricClient) GetRawMetric(metricName string, namespace string, selector labels.Selector, metricSelector labels.Selector) (metricsclient.PodMetricsInfo, time.Time, error) {
	return f.GetRawMetricReactor(metricName, namespace, selector, metricSelector)
}

// GetObjectMetric calls the fake MetricClient function
func (f *MetricClient) GetObjectMetric(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (int64, time.Time, error) {
	return f.GetObjectMetricReactor(metricName, namespace, objectRef, metricSelector)
}

// GetExternalMetric calls the fake MetricClient function
func (f *MetricClient) GetExternalMetric(metricName string, namespace string, selector labels.Selector) ([]int64, time.Time, error) {
	return f.GetExternalMetricReactor(metricName, namespace, selector)
}
