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

package podutil_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/podutil"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	corelisters "k8s.io/client-go/listers/core/v1"
	metricsclient "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

func TestPodReadyCount_GetReadyPodsCount(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    int64
		expectedErr error
		podLister   corelisters.PodLister
		namespace   string
		selector    labels.Selector
	}{
		{
			"Fail to get pods",
			0,
			errors.New("unable to get pods while calculating replica count: fail to get pods"),
			&fake.PodLister{
				PodsReactor: func(namespace string) corelisters.PodNamespaceLister {
					return &fake.PodNamespaceLister{
						ListReactor: func(selector labels.Selector) (ret []*corev1.Pod, err error) {
							return nil, errors.New("fail to get pods")
						},
					}
				},
			},
			"test-namespace",
			nil,
		},
		{
			"0 pods, success",
			0,
			nil,
			&fake.PodLister{
				PodsReactor: func(namespace string) corelisters.PodNamespaceLister {
					return &fake.PodNamespaceLister{
						ListReactor: func(selector labels.Selector) (ret []*corev1.Pod, err error) {
							return []*corev1.Pod{}, nil
						},
					}
				},
			},
			"test-namespace",
			nil,
		},
		{
			"1 ready pod, success",
			1,
			nil,
			&fake.PodLister{
				PodsReactor: func(namespace string) corelisters.PodNamespaceLister {
					return &fake.PodNamespaceLister{
						ListReactor: func(selector labels.Selector) (ret []*corev1.Pod, err error) {
							return []*corev1.Pod{
								{
									Status: corev1.PodStatus{
										Phase: corev1.PodRunning,
										Conditions: []corev1.PodCondition{
											{
												Type:   corev1.PodReady,
												Status: corev1.ConditionTrue,
											},
										},
									},
								},
							}, nil
						},
					}
				},
			},
			"test-namespace",
			nil,
		},
		{
			"1 not ready pod, success",
			0,
			nil,
			&fake.PodLister{
				PodsReactor: func(namespace string) corelisters.PodNamespaceLister {
					return &fake.PodNamespaceLister{
						ListReactor: func(selector labels.Selector) (ret []*corev1.Pod, err error) {
							return []*corev1.Pod{
								{
									Status: corev1.PodStatus{
										Phase: corev1.PodRunning,
										Conditions: []corev1.PodCondition{
											{
												Type:   corev1.PodReady,
												Status: corev1.ConditionFalse,
											},
										},
									},
								},
							}, nil
						},
					}
				},
			},
			"test-namespace",
			nil,
		},
		{
			"2 ready pods, 2 not ready pods, success",
			2,
			nil,
			&fake.PodLister{
				PodsReactor: func(namespace string) corelisters.PodNamespaceLister {
					return &fake.PodNamespaceLister{
						ListReactor: func(selector labels.Selector) (ret []*corev1.Pod, err error) {
							return []*corev1.Pod{
								{
									Status: corev1.PodStatus{
										Phase: corev1.PodRunning,
										Conditions: []corev1.PodCondition{
											{
												Type:   corev1.PodReady,
												Status: corev1.ConditionTrue,
											},
										},
									},
								},
								{
									Status: corev1.PodStatus{
										Phase: corev1.PodRunning,
										Conditions: []corev1.PodCondition{
											{
												Type:   corev1.PodReady,
												Status: corev1.ConditionTrue,
											},
										},
									},
								},
								{
									Status: corev1.PodStatus{
										Phase: corev1.PodRunning,
										Conditions: []corev1.PodCondition{
											{
												Type:   corev1.PodReady,
												Status: corev1.ConditionFalse,
											},
										},
									},
								},
								{
									Status: corev1.PodStatus{
										Phase: corev1.PodRunning,
										Conditions: []corev1.PodCondition{
											{
												Type:   corev1.PodReady,
												Status: corev1.ConditionFalse,
											},
										},
									},
								},
							}, nil
						},
					}
				},
			},
			"test-namespace",
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			podReadyCounter := &podutil.PodReadyCount{
				PodLister: test.podLister,
			}
			readyPods, err := podReadyCounter.GetReadyPodsCount(test.namespace, test.selector)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, readyPods) {
				t.Errorf("ready pods mismatch (-want +got):\n%s", cmp.Diff(test.expected, readyPods))
			}
		})
	}
}

func TestGroupPods(t *testing.T) {
	tests := []struct {
		name                string
		pods                []*corev1.Pod
		metrics             metricsclient.PodMetricsInfo
		resource            corev1.ResourceName
		expectReadyPodCount int
		expectIgnoredPods   sets.String
		expectMissingPods   sets.String
	}{
		{
			"void",
			[]*corev1.Pod{},
			metricsclient.PodMetricsInfo{},
			corev1.ResourceCPU,
			0,
			sets.NewString(),
			sets.NewString(),
		},
		{
			"count in a ready pod - memory",
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bentham",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
					},
				},
			},
			metricsclient.PodMetricsInfo{
				"bentham": metricsclient.PodMetric{Value: 1, Timestamp: time.Now(), Window: time.Minute},
			},
			corev1.ResourceMemory,
			1,
			sets.NewString(),
			sets.NewString(),
		},
		{
			"ignore a pod without ready condition - CPU",
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "lucretius",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
						StartTime: &metav1.Time{
							Time: time.Now(),
						},
					},
				},
			},
			metricsclient.PodMetricsInfo{
				"lucretius": metricsclient.PodMetric{Value: 1},
			},
			corev1.ResourceCPU,
			0,
			sets.NewString("lucretius"),
			sets.NewString(),
		},
		{
			"count in a ready pod with fresh metrics during initialization period - CPU",
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bentham",
					},
					Status: v1.PodStatus{
						Phase: corev1.PodSucceeded,
						StartTime: &metav1.Time{
							Time: time.Now().Add(-1 * time.Minute),
						},
						Conditions: []corev1.PodCondition{
							{
								Type:               corev1.PodReady,
								LastTransitionTime: metav1.Time{Time: time.Now().Add(-30 * time.Second)},
								Status:             corev1.ConditionTrue,
							},
						},
					},
				},
			},
			metricsclient.PodMetricsInfo{
				"bentham": metricsclient.PodMetric{Value: 1, Timestamp: time.Now(), Window: 30 * time.Second},
			},
			corev1.ResourceCPU,
			1,
			sets.NewString(),
			sets.NewString(),
		},
		{
			"ignore a ready pod without fresh metrics during initialization period - CPU",
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bentham",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
						StartTime: &metav1.Time{
							Time: time.Now().Add(-1 * time.Minute),
						},
						Conditions: []corev1.PodCondition{
							{
								Type:               corev1.PodReady,
								LastTransitionTime: metav1.Time{Time: time.Now().Add(-30 * time.Second)},
								Status:             corev1.ConditionTrue,
							},
						},
					},
				},
			},
			metricsclient.PodMetricsInfo{
				"bentham": metricsclient.PodMetric{Value: 1, Timestamp: time.Now(), Window: 60 * time.Second},
			},
			corev1.ResourceCPU,
			0,
			sets.NewString("bentham"),
			sets.NewString(),
		},
		{
			"ignore an unready pod during initialization period - CPU",
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "lucretius",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
						StartTime: &metav1.Time{
							Time: time.Now().Add(-10 * time.Minute),
						},
						Conditions: []corev1.PodCondition{
							{
								Type:               corev1.PodReady,
								LastTransitionTime: metav1.Time{Time: time.Now().Add(-9*time.Minute - 54*time.Second)},
								Status:             corev1.ConditionFalse,
							},
						},
					},
				},
			},
			metricsclient.PodMetricsInfo{
				"lucretius": metricsclient.PodMetric{Value: 1},
			},
			v1.ResourceCPU,
			0,
			sets.NewString("lucretius"),
			sets.NewString(),
		},
		{
			"count in a ready pod without fresh metrics after initialization period - CPU",
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bentham",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
						StartTime: &metav1.Time{
							Time: time.Now().Add(-3 * time.Minute),
						},
						Conditions: []v1.PodCondition{
							{
								Type:               corev1.PodReady,
								LastTransitionTime: metav1.Time{Time: time.Now().Add(-3 * time.Minute)},
								Status:             corev1.ConditionTrue,
							},
						},
					},
				},
			},
			metricsclient.PodMetricsInfo{
				"bentham": metricsclient.PodMetric{Value: 1, Timestamp: time.Now().Add(-2 * time.Minute), Window: time.Minute},
			},
			corev1.ResourceCPU,
			1,
			sets.NewString(),
			sets.NewString(),
		},

		{
			"count in an unready pod that was ready after initialization period - CPU",
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "lucretius",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
						StartTime: &metav1.Time{
							Time: time.Now().Add(-10 * time.Minute),
						},
						Conditions: []corev1.PodCondition{
							{
								Type:               corev1.PodReady,
								LastTransitionTime: metav1.Time{Time: time.Now().Add(-9 * time.Minute)},
								Status:             corev1.ConditionFalse,
							},
						},
					},
				},
			},
			metricsclient.PodMetricsInfo{
				"lucretius": metricsclient.PodMetric{Value: 1},
			},
			corev1.ResourceCPU,
			1,
			sets.NewString(),
			sets.NewString(),
		},
		{
			"ignore pod that has never been ready after initialization period - CPU",
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "lucretius",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
						StartTime: &metav1.Time{
							Time: time.Now().Add(-10 * time.Minute),
						},
						Conditions: []corev1.PodCondition{
							{
								Type:               corev1.PodReady,
								LastTransitionTime: metav1.Time{Time: time.Now().Add(-9*time.Minute - 50*time.Second)},
								Status:             corev1.ConditionFalse,
							},
						},
					},
				},
			},
			metricsclient.PodMetricsInfo{
				"lucretius": metricsclient.PodMetric{Value: 1},
			},
			corev1.ResourceCPU,
			1,
			sets.NewString(),
			sets.NewString(),
		},
		{
			"a missing pod",
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "epicurus",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
						StartTime: &metav1.Time{
							Time: time.Now().Add(-3 * time.Minute),
						},
					},
				},
			},
			metricsclient.PodMetricsInfo{},
			v1.ResourceCPU,
			0,
			sets.NewString(),
			sets.NewString("epicurus"),
		},
		{
			"several pods",
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "lucretius",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
						StartTime: &metav1.Time{
							Time: time.Now(),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "niccolo",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
						StartTime: &metav1.Time{
							Time: time.Now().Add(-3 * time.Minute),
						},
						Conditions: []corev1.PodCondition{
							{
								Type:               corev1.PodReady,
								LastTransitionTime: metav1.Time{Time: time.Now().Add(-3 * time.Minute)},
								Status:             corev1.ConditionTrue,
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "epicurus",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
						StartTime: &metav1.Time{
							Time: time.Now().Add(-3 * time.Minute),
						},
					},
				},
			},
			metricsclient.PodMetricsInfo{
				"lucretius": metricsclient.PodMetric{Value: 1},
				"niccolo":   metricsclient.PodMetric{Value: 1},
			},
			v1.ResourceCPU,
			1,
			sets.NewString("lucretius"),
			sets.NewString("epicurus"),
		},
		{
			name: "pending pods are ignored",
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "unscheduled",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
					},
				},
			},
			metrics:             metricsclient.PodMetricsInfo{},
			resource:            corev1.ResourceCPU,
			expectReadyPodCount: 0,
			expectIgnoredPods:   sets.NewString("unscheduled"),
			expectMissingPods:   sets.NewString(),
		},
		{
			name: "failed pods are skipped",
			pods: []*corev1.Pod{
				{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				},
			},
			metrics:             metricsclient.PodMetricsInfo{},
			resource:            corev1.ResourceCPU,
			expectReadyPodCount: 0,
			expectIgnoredPods:   sets.NewString(),
			expectMissingPods:   sets.NewString(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			readyPodCount, ignoredPods, missingPods := podutil.GroupPods(tc.pods, tc.metrics, tc.resource, 2*time.Minute, 10*time.Second)
			if readyPodCount != tc.expectReadyPodCount {
				t.Errorf("%s got readyPodCount %d, expected %d", tc.name, readyPodCount, tc.expectReadyPodCount)
			}
			if !ignoredPods.Equal(tc.expectIgnoredPods) {
				t.Errorf("%s got unreadyPods %v, expected %v", tc.name, ignoredPods, tc.expectIgnoredPods)
			}
			if !missingPods.Equal(tc.expectMissingPods) {
				t.Errorf("%s got missingPods %v, expected %v", tc.name, missingPods, tc.expectMissingPods)
			}
		})
	}
}

func TestCalculatePodRequests(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})
	var tests = []struct {
		description string
		expected    map[string]int64
		expectedErr error
		pods        []*corev1.Pod
		resource    corev1.ResourceName
	}{
		{
			"Fail missing requests",
			nil,
			errors.New("missing request for test resource"),
			[]*corev1.Pod{
				{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{},
								},
							},
						},
					},
				},
			},
			"test resource",
		},
		{
			"1 pod success",
			map[string]int64{
				"test-pod": 10,
			},
			nil,
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"test resource": *resource.NewMilliQuantity(10, resource.DecimalSI),
									},
								},
							},
						},
					},
				},
			},
			"test resource",
		},
		{
			"3 pod success",
			map[string]int64{
				"test-pod-1": 10,
				"test-pod-2": 20,
				"test-pod-3": 25,
			},
			nil,
			[]*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod-1",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"test resource": *resource.NewMilliQuantity(10, resource.DecimalSI),
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod-2",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"test resource": *resource.NewMilliQuantity(20, resource.DecimalSI),
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod-3",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"test resource": *resource.NewMilliQuantity(20, resource.DecimalSI),
									},
								},
							},
							{
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"test resource": *resource.NewMilliQuantity(5, resource.DecimalSI),
									},
								},
							},
						},
					},
				},
			},
			"test resource",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := podutil.CalculatePodRequests(test.pods, test.resource)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("requests mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestRemoveMetricsForPods(t *testing.T) {
	var tests = []struct {
		description string
		expected    metricsclient.PodMetricsInfo
		metrics     metricsclient.PodMetricsInfo
		pods        sets.String
	}{
		{
			"No pods to remove",
			metricsclient.PodMetricsInfo{
				"test": metricsclient.PodMetric{},
			},
			metricsclient.PodMetricsInfo{
				"test": metricsclient.PodMetric{},
			},
			nil,
		},
		{
			"Remove 3 pods, leave 2",
			metricsclient.PodMetricsInfo{
				"test3": metricsclient.PodMetric{},
				"test4": metricsclient.PodMetric{},
			},
			metricsclient.PodMetricsInfo{
				"test":  metricsclient.PodMetric{},
				"test1": metricsclient.PodMetric{},
				"test2": metricsclient.PodMetric{},
				"test3": metricsclient.PodMetric{},
				"test4": metricsclient.PodMetric{},
			},
			sets.String{
				"test":  {},
				"test1": {},
				"test2": {},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			podutil.RemoveMetricsForPods(test.metrics, test.pods)
			if !cmp.Equal(test.expected, test.metrics) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, test.metrics))
			}
		})
	}
}
