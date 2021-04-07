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

package pods_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/pods"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	corelisters "k8s.io/client-go/listers/core/v1"
	metricsclient "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

func TestGetMetric(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description    string
		expected       *pods.Metric
		expectedErr    error
		metricsClient  metricsclient.MetricsClient
		podLister      corelisters.PodLister
		metricName     string
		namespace      string
		selector       labels.Selector
		metricSelector labels.Selector
	}{
		{
			"Fail to get metric",
			nil,
			errors.New("unable to get metric test-metric: fail to get metric"),
			&fake.MetricClient{
				GetRawMetricReactor: func(metricName, namespace string, selector, metricSelector labels.Selector) (metricsclient.PodMetricsInfo, time.Time, error) {
					return nil, time.Time{}, errors.New("fail to get metric")
				},
			},
			nil,
			"test-metric",
			"test-namespace",
			nil,
			nil,
		},
		{
			"Fail to get pods",
			nil,
			errors.New("unable to get pods while calculating replica count: fail to get pods"),
			&fake.MetricClient{
				GetRawMetricReactor: func(metricName, namespace string, selector, metricSelector labels.Selector) (metricsclient.PodMetricsInfo, time.Time, error) {
					return nil, time.Time{}, nil
				},
			},
			&fake.PodLister{
				PodsReactor: func(namespace string) corelisters.PodNamespaceLister {
					return &fake.PodNamespaceLister{
						ListReactor: func(selector labels.Selector) (ret []*v1.Pod, err error) {
							return nil, errors.New("fail to get pods")
						},
					}
				},
			},
			"test-metric",
			"test-namespace",
			nil,
			nil,
		},
		{
			"No pods success",
			&pods.Metric{
				ReadyPodCount: 0,
				TotalPods:     0,
				Timestamp:     time.Time{},
			},
			nil,
			&fake.MetricClient{
				GetRawMetricReactor: func(metricName, namespace string, selector, metricSelector labels.Selector) (metricsclient.PodMetricsInfo, time.Time, error) {
					return metricsclient.PodMetricsInfo{
						"test-pod": metricsclient.PodMetric{},
					}, time.Time{}, nil
				},
			},
			&fake.PodLister{
				PodsReactor: func(namespace string) corelisters.PodNamespaceLister {
					return &fake.PodNamespaceLister{
						ListReactor: func(selector labels.Selector) (ret []*v1.Pod, err error) {
							return []*v1.Pod{}, nil
						},
					}
				},
			},
			"test-metric",
			"test-namespace",
			nil,
			nil,
		},
		{
			"3 ready, 2 missing pods success",
			&pods.Metric{
				TotalPods:     5,
				ReadyPodCount: 3,
				MissingPods: sets.String{
					"missing-pod-1": {},
					"missing-pod-2": {},
				},
				PodMetricsInfo: metricsclient.PodMetricsInfo{
					"ready-pod-1": metricsclient.PodMetric{},
					"ready-pod-2": metricsclient.PodMetric{},
					"ready-pod-3": metricsclient.PodMetric{},
				},
			},
			nil,
			&fake.MetricClient{
				GetRawMetricReactor: func(metricName, namespace string, selector, metricSelector labels.Selector) (metricsclient.PodMetricsInfo, time.Time, error) {
					return metricsclient.PodMetricsInfo{
						"ready-pod-1": metricsclient.PodMetric{},
						"ready-pod-2": metricsclient.PodMetric{},
						"ready-pod-3": metricsclient.PodMetric{},
					}, time.Time{}, nil
				},
			},
			&fake.PodLister{
				PodsReactor: func(namespace string) corelisters.PodNamespaceLister {
					return &fake.PodNamespaceLister{
						ListReactor: func(selector labels.Selector) (ret []*v1.Pod, err error) {
							return []*v1.Pod{
								{
									ObjectMeta: metav1.ObjectMeta{
										Name: "ready-pod-1",
									},
								},
								{
									ObjectMeta: metav1.ObjectMeta{
										Name: "ready-pod-2",
									},
								},
								{
									ObjectMeta: metav1.ObjectMeta{
										Name: "ready-pod-3",
									},
								},
								{
									ObjectMeta: metav1.ObjectMeta{
										Name: "missing-pod-1",
									},
								},
								{
									ObjectMeta: metav1.ObjectMeta{
										Name: "missing-pod-2",
									},
								},
							}, nil
						},
					}
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
			gatherer := &pods.Gather{
				MetricsClient: test.metricsClient,
				PodLister:     test.podLister,
			}
			metric, err := gatherer.GetMetric(test.metricName, test.namespace, test.selector, test.metricSelector)
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
