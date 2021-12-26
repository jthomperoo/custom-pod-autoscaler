//go:build unit
// +build unit

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

package externalget_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/externalget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/podutil"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/external"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/value"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		expected        *external.Metric
		expectedErr     error
		metricsClient   metricsclient.MetricsClient
		podReadyCounter podutil.PodReadyCounter
		metricName      string
		namespace       string
		metricSelector  *metav1.LabelSelector
		podSelector     labels.Selector
	}{
		{
			"Fail convert metric selector",
			nil,
			errors.New(`"invalid" is not a valid pod selector operator`),
			nil,
			nil,
			"test-metric",
			"test-namespace",
			&metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Operator: "invalid",
					},
				},
			},
			nil,
		},
		{
			"Fail to get metric",
			nil,
			errors.New("unable to get external metric test-namespace/test-metric/nil: fail to get metric"),
			&fake.MetricClient{
				GetExternalMetricReactor: func(metricName, namespace string, selector labels.Selector) ([]int64, time.Time, error) {
					return []int64{}, time.Time{}, errors.New("fail to get metric")
				},
			},
			nil,
			"test-metric",
			"test-namespace",
			nil,
			nil,
		},
		{
			"Fail to get ready pods",
			nil,
			errors.New("unable to calculate ready pods: fail to get ready pods"),
			&fake.MetricClient{
				GetExternalMetricReactor: func(metricName, namespace string, selector labels.Selector) ([]int64, time.Time, error) {
					return []int64{}, time.Time{}, nil
				},
			},
			&fake.PodReadyCounter{
				GetReadyPodsCountReactor: func(namespace string, selector labels.Selector) (int64, error) {
					return 0, errors.New("fail to get ready pods")
				},
			},
			"test-metric",
			"test-namespace",
			nil,
			nil,
		},
		{
			"5 ready pods, 5 metrics, success",
			&external.Metric{
				ReadyPodCount: int64Ptr(5),
				Current: value.MetricValue{
					Value: int64Ptr(15),
				},
			},
			nil,
			&fake.MetricClient{
				GetExternalMetricReactor: func(metricName, namespace string, selector labels.Selector) ([]int64, time.Time, error) {
					return []int64{1, 2, 3, 4, 5}, time.Time{}, nil
				},
			},
			&fake.PodReadyCounter{
				GetReadyPodsCountReactor: func(namespace string, selector labels.Selector) (int64, error) {
					return 5, nil
				},
			},
			"test-metric",
			"test-namespace",
			nil,
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := &externalget.Gather{
				MetricsClient:   test.metricsClient,
				PodReadyCounter: test.podReadyCounter,
			}
			metric, err := gatherer.GetMetric(test.metricName, test.namespace, test.metricSelector, test.podSelector)
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
		expected        *external.Metric
		expectedErr     error
		metricsClient   metricsclient.MetricsClient
		podReadyCounter podutil.PodReadyCounter
		metricName      string
		namespace       string
		metricSelector  *metav1.LabelSelector
	}{
		{
			"Fail convert metric selector",
			nil,
			errors.New(`"invalid" is not a valid pod selector operator`),
			nil,
			nil,
			"test-metric",
			"test-namespace",
			&metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Operator: "invalid",
					},
				},
			},
		},
		{
			"Fail to get metric",
			nil,
			errors.New("unable to get external metric test-namespace/test-metric/nil: fail to get metric"),
			&fake.MetricClient{
				GetExternalMetricReactor: func(metricName, namespace string, selector labels.Selector) ([]int64, time.Time, error) {
					return []int64{}, time.Time{}, errors.New("fail to get metric")
				},
			},
			nil,
			"test-metric",
			"test-namespace",
			nil,
		},
		{
			"5 metrics, success",
			&external.Metric{
				Current: value.MetricValue{
					AverageValue: int64Ptr(15),
				},
			},
			nil,
			&fake.MetricClient{
				GetExternalMetricReactor: func(metricName, namespace string, selector labels.Selector) ([]int64, time.Time, error) {
					return []int64{1, 2, 3, 4, 5}, time.Time{}, nil
				},
			},
			nil,
			"test-metric",
			"test-namespace",
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := &externalget.Gather{
				MetricsClient:   test.metricsClient,
				PodReadyCounter: test.podReadyCounter,
			}
			metric, err := gatherer.GetPerPodMetric(test.metricName, test.namespace, test.metricSelector)
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
