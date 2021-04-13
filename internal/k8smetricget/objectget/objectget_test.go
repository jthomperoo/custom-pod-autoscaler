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

package objectget_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/objectget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/podutil"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/object"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/value"
	"k8s.io/api/autoscaling/v2beta2"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/labels"
	metricsclient "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

func int64Ptr(i int64) *int64 {
	return &i
}

func TestGetMetric(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description     string
		expected        *object.Metric
		expectedErr     error
		metricsClient   metricsclient.MetricsClient
		podReadyCounter podutil.PodReadyCounter
		metricName      string
		namespace       string
		objectRef       *v2beta2.CrossVersionObjectReference
		selector        labels.Selector
		metricSelector  labels.Selector
	}{
		{
			"Fail to get metric",
			nil,
			errors.New("unable to get metric test-metric:  on test-namespace /fail to get metric"),
			&fake.MetricClient{
				GetObjectMetricReactor: func(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (int64, time.Time, error) {
					return 0, time.Time{}, errors.New("fail to get metric")
				},
			},
			nil,
			"test-metric",
			"test-namespace",
			&autoscaling.CrossVersionObjectReference{},
			nil,
			nil,
		},
		{
			"Fail to get ready pods",
			nil,
			errors.New("unable to calculate ready pods: fail to get ready pods"),
			&fake.MetricClient{
				GetObjectMetricReactor: func(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (int64, time.Time, error) {
					return 0, time.Time{}, nil
				},
			},
			&fake.PodReadyCounter{
				GetReadyPodsCountReactor: func(namespace string, selector labels.Selector) (int64, error) {
					return 0, errors.New("fail to get ready pods")
				},
			},
			"test-metric",
			"test-namespace",
			&autoscaling.CrossVersionObjectReference{},
			nil,
			nil,
		},
		{
			"Success",
			&object.Metric{
				Current: value.MetricValue{
					Value: int64Ptr(5),
				},
				ReadyPodCount: int64Ptr(2),
			},
			nil,
			&fake.MetricClient{
				GetObjectMetricReactor: func(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (int64, time.Time, error) {
					return 5, time.Time{}, nil
				},
			},
			&fake.PodReadyCounter{
				GetReadyPodsCountReactor: func(namespace string, selector labels.Selector) (int64, error) {
					return 2, nil
				},
			},
			"test-metric",
			"test-namespace",
			&autoscaling.CrossVersionObjectReference{},
			nil,
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := &objectget.Gather{
				MetricsClient:   test.metricsClient,
				PodReadyCounter: test.podReadyCounter,
			}
			metric, err := gatherer.GetMetric(test.metricName, test.namespace, test.objectRef, test.selector, test.metricSelector)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, metric) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, metric))
			}
		})
	}
}

func TestGetPerPodMetric(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description     string
		expected        *object.Metric
		expectedErr     error
		metricsClient   metricsclient.MetricsClient
		podReadyCounter podutil.PodReadyCounter
		metricName      string
		namespace       string
		objectRef       *v2beta2.CrossVersionObjectReference
		metricSelector  labels.Selector
	}{
		{
			"Fail to get metric",
			nil,
			errors.New("unable to get metric test-metric:  on test-namespace /fail to get metric"),
			&fake.MetricClient{
				GetObjectMetricReactor: func(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (int64, time.Time, error) {
					return 0, time.Time{}, errors.New("fail to get metric")
				},
			},
			nil,
			"test-metric",
			"test-namespace",
			&autoscaling.CrossVersionObjectReference{},
			nil,
		},
		{
			"Success",
			&object.Metric{
				Current: value.MetricValue{
					AverageValue: int64Ptr(5),
				},
			},
			nil,
			&fake.MetricClient{
				GetObjectMetricReactor: func(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (int64, time.Time, error) {
					return 5, time.Time{}, nil
				},
			},
			nil,
			"test-metric",
			"test-namespace",
			&autoscaling.CrossVersionObjectReference{},
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := &objectget.Gather{
				MetricsClient:   test.metricsClient,
				PodReadyCounter: test.podReadyCounter,
			}
			metric, err := gatherer.GetPerPodMetric(test.metricName, test.namespace, test.objectRef, test.metricSelector)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, metric) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, metric))
			}
		})
	}
}
