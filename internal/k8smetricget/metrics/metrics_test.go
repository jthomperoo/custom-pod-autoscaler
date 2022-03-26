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

package metrics_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/metrics"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/podmetrics"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
	custom_metricsv1beta2 "k8s.io/metrics/pkg/apis/custom_metrics/v1beta2"
	external_metricsv1beta1 "k8s.io/metrics/pkg/apis/external_metrics/v1beta1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv1beta1fake "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1/fake"
	custom_metricsfake "k8s.io/metrics/pkg/client/custom_metrics/fake"
	external_metricsfake "k8s.io/metrics/pkg/client/external_metrics/fake"
)

func int64Ptr(i int64) *int64 {
	return &i
}

func TestGetResourceMetric(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description  string
		expectedInfo podmetrics.MetricsInfo
		expectedTime time.Time
		expectedErr  error
		client       metrics.RESTClient
		resource     v1.ResourceName
		namespace    string
		selector     labels.Selector
	}{
		{
			description:  "Fail, fail to fetch metrics",
			expectedInfo: nil,
			expectedTime: time.Time{},
			expectedErr:  errors.New("unable to fetch metrics from resource metrics API: fail to get pod metrics"),
			client: metrics.RESTClient{
				Client: &metricsv1beta1fake.FakeMetricsV1beta1{
					Fake: &k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "pods",
								Verb:     "list",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, nil, errors.New("fail to get pod metrics")
								},
							},
						},
					},
				},
			},
			resource:  v1.ResourceCPU,
			namespace: "test",
			selector:  labels.Everything(),
		},
		{
			description:  "Fail, no metrics found",
			expectedInfo: nil,
			expectedTime: time.Time{},
			expectedErr:  errors.New("no metrics returned from resource metrics API"),
			client: metrics.RESTClient{
				Client: &metricsv1beta1fake.FakeMetricsV1beta1{
					Fake: &k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "pods",
								Verb:     "list",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &metricsv1beta1.PodMetricsList{
										Items: []metricsv1beta1.PodMetrics{},
									}, nil
								},
							},
						},
					},
				},
			},
			resource:  v1.ResourceCPU,
			namespace: "test",
			selector:  labels.Everything(),
		},
		{
			description:  "Success, one metric, no containers",
			expectedInfo: podmetrics.MetricsInfo{},
			expectedTime: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
			expectedErr:  nil,
			client: metrics.RESTClient{
				Client: &metricsv1beta1fake.FakeMetricsV1beta1{
					Fake: &k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "pods",
								Verb:     "list",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &metricsv1beta1.PodMetricsList{
										Items: []metricsv1beta1.PodMetrics{
											{
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
												},
											},
										},
									}, nil
								},
							},
						},
					},
				},
			},
			resource:  v1.ResourceCPU,
			namespace: "test",
			selector:  labels.Everything(),
		},
		{
			description:  "Success, one metric, one container, desired metric not found",
			expectedInfo: podmetrics.MetricsInfo{},
			expectedTime: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
			expectedErr:  nil,
			client: metrics.RESTClient{
				Client: &metricsv1beta1fake.FakeMetricsV1beta1{
					Fake: &k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "pods",
								Verb:     "list",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &metricsv1beta1.PodMetricsList{
										Items: []metricsv1beta1.PodMetrics{
											{
												ObjectMeta: metav1.ObjectMeta{
													Name: "test",
												},
												Containers: []metricsv1beta1.ContainerMetrics{
													{
														Usage: v1.ResourceList{
															v1.ResourceName("test"): *resource.NewQuantity(10, resource.DecimalSI),
														},
													},
												},
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
												},
											},
										},
									}, nil
								},
							},
						},
					},
				},
			},
			resource:  v1.ResourceCPU,
			namespace: "test",
			selector:  labels.Everything(),
		},
		{
			description: "Success, one metric, one container, desired metric found",
			expectedInfo: podmetrics.MetricsInfo{
				"test": podmetrics.Metric{
					Timestamp: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
					Value:     10000,
				},
			},
			expectedTime: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
			expectedErr:  nil,
			client: metrics.RESTClient{
				Client: &metricsv1beta1fake.FakeMetricsV1beta1{
					Fake: &k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "pods",
								Verb:     "list",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &metricsv1beta1.PodMetricsList{
										Items: []metricsv1beta1.PodMetrics{
											{
												ObjectMeta: metav1.ObjectMeta{
													Name: "test",
												},
												Containers: []metricsv1beta1.ContainerMetrics{
													{
														Usage: v1.ResourceList{
															v1.ResourceCPU: *resource.NewQuantity(10, resource.DecimalSI),
														},
													},
												},
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
												},
											},
										},
									}, nil
								},
							},
						},
					},
				},
			},
			resource:  v1.ResourceCPU,
			namespace: "test",
			selector:  labels.Everything(),
		},
		{
			description: "Success, three metrics, two containers each, desired metric found",
			expectedInfo: podmetrics.MetricsInfo{
				"test1": podmetrics.Metric{
					Timestamp: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
					Value:     10000,
				},
				"test3": podmetrics.Metric{
					Timestamp: time.Date(2000, 3, 7, 10, 30, 0, 5, time.UTC),
					Value:     15000,
				},
			},
			expectedTime: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
			expectedErr:  nil,
			client: metrics.RESTClient{
				Client: &metricsv1beta1fake.FakeMetricsV1beta1{
					Fake: &k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "pods",
								Verb:     "list",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &metricsv1beta1.PodMetricsList{
										Items: []metricsv1beta1.PodMetrics{
											{
												ObjectMeta: metav1.ObjectMeta{
													Name: "test1",
												},
												Containers: []metricsv1beta1.ContainerMetrics{
													{
														Usage: v1.ResourceList{
															v1.ResourceCPU:    *resource.NewQuantity(10, resource.DecimalSI),
															v1.ResourceMemory: *resource.NewQuantity(20, resource.DecimalSI),
														},
													},
												},
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
												},
											},
											{
												ObjectMeta: metav1.ObjectMeta{
													Name: "test2",
												},
												Containers: []metricsv1beta1.ContainerMetrics{
													{
														Usage: v1.ResourceList{
															v1.ResourceName("test"): *resource.NewQuantity(3, resource.DecimalSI),
														},
													},
												},
												Timestamp: metav1.Time{
													Time: time.Date(1999, 3, 7, 10, 30, 0, 5, time.UTC),
												},
											},
											{
												ObjectMeta: metav1.ObjectMeta{
													Name: "test3",
												},
												Containers: []metricsv1beta1.ContainerMetrics{
													{
														Usage: v1.ResourceList{
															v1.ResourceCPU: *resource.NewQuantity(15, resource.DecimalSI),
														},
													},
												},
												Timestamp: metav1.Time{
													Time: time.Date(2000, 3, 7, 10, 30, 0, 5, time.UTC),
												},
											},
										},
									}, nil
								},
							},
						},
					},
				},
			},
			resource:  v1.ResourceCPU,
			namespace: "test",
			selector:  labels.Everything(),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			info, time, err := test.client.GetResourceMetric(test.resource, test.namespace, test.selector)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expectedInfo, info) {
				t.Errorf("info mismatch (-want +got):\n%s", cmp.Diff(test.expectedInfo, info))
			}
			if !cmp.Equal(test.expectedTime, time) {
				t.Errorf("time mismatch (-want +got):\n%s", cmp.Diff(test.expectedTime, time))
			}
		})
	}
}

func Test_GetRawMetric(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description    string
		expectedInfo   podmetrics.MetricsInfo
		expectedTime   time.Time
		expectedErr    error
		client         metrics.RESTClient
		metricName     string
		namespace      string
		selector       labels.Selector
		metricSelector labels.Selector
	}{
		{
			description:  "Fail, fail to fetch metrics",
			expectedInfo: nil,
			expectedTime: time.Time{},
			expectedErr:  errors.New("unable to fetch metrics from custom metrics API: fail to get pod metrics"),
			client: metrics.RESTClient{
				CustomMetricsClient: &custom_metricsfake.FakeCustomMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "pods",
								Verb:     "get",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, nil, errors.New("fail to get pod metrics")
								},
							},
						},
					},
				},
			},
			metricName:     "test",
			namespace:      "test",
			selector:       labels.Everything(),
			metricSelector: labels.Everything(),
		},
		{
			description:  "Fail, no metrics returned",
			expectedInfo: nil,
			expectedTime: time.Time{},
			expectedErr:  errors.New("no metrics returned from custom metrics API"),
			client: metrics.RESTClient{
				CustomMetricsClient: &custom_metricsfake.FakeCustomMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "pods",
								Verb:     "get",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &custom_metricsv1beta2.MetricValueList{}, nil
								},
							},
						},
					},
				},
			},
			metricName:     "test",
			namespace:      "test",
			selector:       labels.Everything(),
			metricSelector: labels.Everything(),
		},
		{
			description: "Success, single metric",
			expectedInfo: podmetrics.MetricsInfo{
				"test": podmetrics.Metric{
					Timestamp: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
					Value:     10000,
					Window:    time.Second * 90,
				},
			},
			expectedTime: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
			expectedErr:  nil,
			client: metrics.RESTClient{
				CustomMetricsClient: &custom_metricsfake.FakeCustomMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "pods",
								Verb:     "get",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &custom_metricsv1beta2.MetricValueList{
										Items: []custom_metricsv1beta2.MetricValue{
											{
												DescribedObject: v1.ObjectReference{
													Name: "test",
												},
												WindowSeconds: int64Ptr(90),
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
												},
												Value: *resource.NewQuantity(10, resource.DecimalSI),
											},
										},
									}, nil
								},
							},
						},
					},
				},
			},
			metricName:     "test",
			namespace:      "test",
			selector:       labels.Everything(),
			metricSelector: labels.Everything(),
		},
		{
			description: "Success, single metric, window not provided",
			expectedInfo: podmetrics.MetricsInfo{
				"test": podmetrics.Metric{
					Timestamp: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
					Value:     10000,
					Window:    time.Second * 60,
				},
			},
			expectedTime: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
			expectedErr:  nil,
			client: metrics.RESTClient{
				CustomMetricsClient: &custom_metricsfake.FakeCustomMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "pods",
								Verb:     "get",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &custom_metricsv1beta2.MetricValueList{
										Items: []custom_metricsv1beta2.MetricValue{
											{
												DescribedObject: v1.ObjectReference{
													Name: "test",
												},
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
												},
												Value: *resource.NewQuantity(10, resource.DecimalSI),
											},
										},
									}, nil
								},
							},
						},
					},
				},
			},
			metricName:     "test",
			namespace:      "test",
			selector:       labels.Everything(),
			metricSelector: labels.Everything(),
		},
		{
			description: "Success, multiple metric",
			expectedInfo: podmetrics.MetricsInfo{
				"test1": podmetrics.Metric{
					Timestamp: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
					Value:     10000,
					Window:    time.Second * 60,
				},
				"test2": podmetrics.Metric{
					Timestamp: time.Date(1998, 3, 8, 10, 30, 0, 5, time.UTC),
					Value:     15000,
					Window:    time.Second * 90,
				},
				"test3": podmetrics.Metric{
					Timestamp: time.Date(1998, 3, 9, 10, 30, 0, 5, time.UTC),
					Value:     20000,
					Window:    time.Second * 60,
				},
			},
			expectedTime: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
			expectedErr:  nil,
			client: metrics.RESTClient{
				CustomMetricsClient: &custom_metricsfake.FakeCustomMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "pods",
								Verb:     "get",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &custom_metricsv1beta2.MetricValueList{
										Items: []custom_metricsv1beta2.MetricValue{
											{
												DescribedObject: v1.ObjectReference{
													Name: "test1",
												},
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
												},
												Value: *resource.NewQuantity(10, resource.DecimalSI),
											},
											{
												DescribedObject: v1.ObjectReference{
													Name: "test2",
												},
												WindowSeconds: int64Ptr(90),
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 8, 10, 30, 0, 5, time.UTC),
												},
												Value: *resource.NewQuantity(15, resource.DecimalSI),
											},
											{
												DescribedObject: v1.ObjectReference{
													Name: "test3",
												},
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 9, 10, 30, 0, 5, time.UTC),
												},
												Value: *resource.NewQuantity(20, resource.DecimalSI),
											},
										},
									}, nil
								},
							},
						},
					},
				},
			},
			metricName:     "test",
			namespace:      "test",
			selector:       labels.Everything(),
			metricSelector: labels.Everything(),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			info, time, err := test.client.GetRawMetric(test.metricName, test.namespace, test.selector, test.metricSelector)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expectedInfo, info) {
				t.Errorf("info mismatch (-want +got):\n%s", cmp.Diff(test.expectedInfo, info))
			}
			if !cmp.Equal(test.expectedTime, time) {
				t.Errorf("time mismatch (-want +got):\n%s", cmp.Diff(test.expectedTime, time))
			}
		})
	}
}

func Test_GetObjectMetric(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description    string
		expectedMetric int64
		expectedTime   time.Time
		expectedErr    error
		client         metrics.RESTClient
		metricName     string
		namespace      string
		objectRef      *autoscalingv2.CrossVersionObjectReference
		metricSelector labels.Selector
	}{
		{
			description:    "Fail, fail to fetch namespaced metrics",
			expectedMetric: 0,
			expectedTime:   time.Time{},
			expectedErr:    errors.New("unable to fetch metrics from custom metrics API: fail to get deployment metrics"),
			client: metrics.RESTClient{
				CustomMetricsClient: &custom_metricsfake.FakeCustomMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "*",
								Verb:     "get",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, nil, errors.New("fail to get deployment metrics")
								},
							},
						},
					},
				},
			},
			metricName: "test",
			namespace:  "test",
			objectRef: &autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test",
			},
			metricSelector: labels.Everything(),
		},
		{
			description:    "Fail, fail to fetch root scoped metrics",
			expectedMetric: 0,
			expectedTime:   time.Time{},
			expectedErr:    errors.New("unable to fetch metrics from custom metrics API: fail to get deployment metrics"),
			client: metrics.RESTClient{
				CustomMetricsClient: &custom_metricsfake.FakeCustomMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "*",
								Verb:     "get",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, nil, errors.New("fail to get deployment metrics")
								},
							},
						},
					},
				},
			},
			metricName: "test",
			namespace:  "test",
			objectRef: &autoscalingv2.CrossVersionObjectReference{
				APIVersion: "v1",
				Kind:       "Namespace",
				Name:       "test",
			},
			metricSelector: labels.Everything(),
		},
		{
			description:    "Success, return namespaced metric",
			expectedMetric: 10000,
			expectedTime:   time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
			expectedErr:    nil,
			client: metrics.RESTClient{
				CustomMetricsClient: &custom_metricsfake.FakeCustomMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "*",
								Verb:     "get",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &custom_metricsv1beta2.MetricValueList{
										Items: []custom_metricsv1beta2.MetricValue{
											{
												DescribedObject: v1.ObjectReference{
													Name: "test",
												},
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
												},
												Value: *resource.NewQuantity(10, resource.DecimalSI),
											},
										},
									}, nil
								},
							},
						},
					},
				},
			},
			metricName: "test",
			namespace:  "test",
			objectRef: &autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test",
			},
			metricSelector: labels.Everything(),
		},
		{
			description:    "Success, return root scoped metric",
			expectedMetric: 10000,
			expectedTime:   time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
			expectedErr:    nil,
			client: metrics.RESTClient{
				CustomMetricsClient: &custom_metricsfake.FakeCustomMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "*",
								Verb:     "get",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &custom_metricsv1beta2.MetricValueList{
										Items: []custom_metricsv1beta2.MetricValue{
											{
												DescribedObject: v1.ObjectReference{
													Name: "test",
												},
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
												},
												Value: *resource.NewQuantity(10, resource.DecimalSI),
											},
										},
									}, nil
								},
							},
						},
					},
				},
			},
			metricName: "test",
			namespace:  "test",
			objectRef: &autoscalingv2.CrossVersionObjectReference{
				APIVersion: "v1",
				Kind:       "Namespace",
				Name:       "test",
			},
			metricSelector: labels.Everything(),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			metric, time, err := test.client.GetObjectMetric(test.metricName, test.namespace, test.objectRef, test.metricSelector)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expectedMetric, metric) {
				t.Errorf("info mismatch (-want +got):\n%s", cmp.Diff(test.expectedMetric, metric))
			}
			if !cmp.Equal(test.expectedTime, time) {
				t.Errorf("time mismatch (-want +got):\n%s", cmp.Diff(test.expectedTime, time))
			}
		})
	}
}

func Test_GetExternalMetric(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description     string
		expectedMetrics []int64
		expectedTime    time.Time
		expectedErr     error
		client          metrics.RESTClient
		metricName      string
		namespace       string
		selector        labels.Selector
	}{
		{
			description:     "Fail, fail to fetch metrics",
			expectedMetrics: []int64{},
			expectedTime:    time.Time{},
			expectedErr:     errors.New("unable to fetch metrics from external metrics API: Fail to get external metrics"),
			client: metrics.RESTClient{
				ExternalMetricsClient: &external_metricsfake.FakeExternalMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "*",
								Verb:     "*",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, nil, errors.New("Fail to get external metrics")
								},
							},
						},
					},
				},
			},
			metricName: "test",
			namespace:  "test",
			selector:   labels.Everything(),
		},
		{
			description:     "Fail, no metrics returned",
			expectedMetrics: nil,
			expectedTime:    time.Time{},
			expectedErr:     errors.New("no metrics returned from external metrics API"),
			client: metrics.RESTClient{
				ExternalMetricsClient: &external_metricsfake.FakeExternalMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "*",
								Verb:     "*",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &external_metricsv1beta1.ExternalMetricValueList{}, nil
								},
							},
						},
					},
				},
			},
			metricName: "test",
			namespace:  "test",
			selector:   labels.Everything(),
		},
		{
			description:     "Success, single metric",
			expectedMetrics: []int64{10000},
			expectedTime:    time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
			expectedErr:     nil,
			client: metrics.RESTClient{
				ExternalMetricsClient: &external_metricsfake.FakeExternalMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "*",
								Verb:     "*",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &external_metricsv1beta1.ExternalMetricValueList{
										Items: []external_metricsv1beta1.ExternalMetricValue{
											{
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
												},
												Value: *resource.NewQuantity(10, resource.DecimalSI),
											},
										},
									}, nil
								},
							},
						},
					},
				},
			},
			metricName: "test",
			namespace:  "test",
			selector:   labels.Everything(),
		},
		{
			description:     "Success, single metric",
			expectedMetrics: []int64{10000, 15000, 20000},
			expectedTime:    time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
			expectedErr:     nil,
			client: metrics.RESTClient{
				ExternalMetricsClient: &external_metricsfake.FakeExternalMetricsClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "*",
								Verb:     "*",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &external_metricsv1beta1.ExternalMetricValueList{
										Items: []external_metricsv1beta1.ExternalMetricValue{
											{
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 7, 10, 30, 0, 5, time.UTC),
												},
												Value: *resource.NewQuantity(10, resource.DecimalSI),
											},
											{
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 8, 10, 30, 0, 5, time.UTC),
												},
												Value: *resource.NewQuantity(15, resource.DecimalSI),
											},
											{
												Timestamp: metav1.Time{
													Time: time.Date(1998, 3, 9, 10, 30, 0, 5, time.UTC),
												},
												Value: *resource.NewQuantity(20, resource.DecimalSI),
											},
										},
									}, nil
								},
							},
						},
					},
				},
			},
			metricName: "test",
			namespace:  "test",
			selector:   labels.Everything(),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			metrics, time, err := test.client.GetExternalMetric(test.metricName, test.namespace, test.selector)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expectedMetrics, metrics) {
				t.Errorf("info mismatch (-want +got):\n%s", cmp.Diff(test.expectedMetrics, metrics))
			}
			if !cmp.Equal(test.expectedTime, time) {
				t.Errorf("time mismatch (-want +got):\n%s", cmp.Diff(test.expectedTime, time))
			}
		})
	}
}
