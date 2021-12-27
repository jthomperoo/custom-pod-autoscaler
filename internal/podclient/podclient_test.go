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

package podclient_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/podclient"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestOnDemandPodLister_List(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    []*corev1.Pod
		expectedErr error
		clientset   kubernetes.Interface
		selector    labels.Selector
	}{
		{
			"Error getting pods",
			nil,
			errors.New("Fail to list pods"),
			func() *fake.Clientset {
				clientset := fake.NewSimpleClientset()
				clientset.CoreV1().(*fakecorev1.FakeCoreV1).Fake.PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("Fail to list pods")
				})
				return clientset
			}(),
			labels.NewSelector(),
		},
		{
			"List 1 pod",
			[]*corev1.Pod{
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-namespace",
					},
				},
			},
			nil,
			fake.NewSimpleClientset(&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			}),
			labels.NewSelector(),
		},
		{
			"List 3 pods in one namespace, 2 in another",
			[]*corev1.Pod{
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-1",
						Namespace: "test-namespace-1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-2",
						Namespace: "test-namespace-1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-3",
						Namespace: "test-namespace-1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-4",
						Namespace: "test-namespace-2",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-5",
						Namespace: "test-namespace-2",
					},
				},
			},
			nil,
			fake.NewSimpleClientset(
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-1",
						Namespace: "test-namespace-1",
					},
				},
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-2",
						Namespace: "test-namespace-1",
					},
				},
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-3",
						Namespace: "test-namespace-1",
					},
				},
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-4",
						Namespace: "test-namespace-2",
					},
				},
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-5",
						Namespace: "test-namespace-2",
					},
				}),
			labels.NewSelector(),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := &podclient.OnDemandPodLister{
				Clientset: test.clientset,
			}
			pods, err := gatherer.List(test.selector)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, pods) {
				t.Errorf("pods mismatch (-want +got):\n%s", cmp.Diff(test.expected, pods))
			}
		})
	}
}

func TestOnDemandPodLister_Pods(t *testing.T) {
	var tests = []struct {
		description string
		expected    *podclient.OnDemandPodNamespaceLister
		clientset   kubernetes.Interface
		namespace   string
	}{
		{
			"Success get namespaced pod lister",
			&podclient.OnDemandPodNamespaceLister{
				Namespace: "test-namespace",
				Clientset: nil,
			},
			nil,
			"test-namespace",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := &podclient.OnDemandPodLister{
				Clientset: test.clientset,
			}
			namespaceLister := gatherer.Pods(test.namespace)
			if !cmp.Equal(test.expected, namespaceLister) {
				t.Errorf("namespace lister mismatch (-want +got):\n%s", cmp.Diff(test.expected, namespaceLister))
			}
		})
	}
}

func TestOnDemandPodNamespaceLister_List(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    []*corev1.Pod
		expectedErr error
		clientset   kubernetes.Interface
		namespace   string
		selector    labels.Selector
	}{
		{
			"Error getting pods",
			nil,
			errors.New("Fail to list pods"),
			func() *fake.Clientset {
				clientset := fake.NewSimpleClientset()
				clientset.CoreV1().(*fakecorev1.FakeCoreV1).Fake.PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("Fail to list pods")
				})
				return clientset
			}(),
			"test-namespace",
			labels.NewSelector(),
		},
		{
			"List 1 pod",
			[]*corev1.Pod{
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-namespace",
					},
				},
			},
			nil,
			fake.NewSimpleClientset(&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			}),
			"test-namespace",
			labels.NewSelector(),
		},
		{
			"List 3 pods in requested namespace, 2 in another",
			[]*corev1.Pod{
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-1",
						Namespace: "test-namespace-1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-2",
						Namespace: "test-namespace-1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-3",
						Namespace: "test-namespace-1",
					},
				},
			},
			nil,
			fake.NewSimpleClientset(
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-1",
						Namespace: "test-namespace-1",
					},
				},
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-2",
						Namespace: "test-namespace-1",
					},
				},
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-3",
						Namespace: "test-namespace-1",
					},
				},
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-4",
						Namespace: "test-namespace-2",
					},
				},
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-pod-5",
						Namespace: "test-namespace-2",
					},
				}),
			"test-namespace-1",
			labels.NewSelector(),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := &podclient.OnDemandPodNamespaceLister{
				Namespace: test.namespace,
				Clientset: test.clientset,
			}
			pods, err := gatherer.List(test.selector)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, pods) {
				t.Errorf("pods mismatch (-want +got):\n%s", cmp.Diff(test.expected, pods))
			}
		})
	}
}

func TestOnDemandPodNamespaceLister_Get(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    *corev1.Pod
		expectedErr error
		clientset   kubernetes.Interface
		namespace   string
		name        string
	}{
		{
			"Error getting pod",
			nil,
			errors.New("Fail to get pod"),
			func() *fake.Clientset {
				clientset := fake.NewSimpleClientset()
				clientset.CoreV1().(*fakecorev1.FakeCoreV1).Fake.PrependReactor("get", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("Fail to get pod")
				})
				return clientset
			}(),
			"test-namespace",
			"test",
		},
		{
			"Don't retrieve pod in other namespace",
			nil,
			errors.New(`pods "test-pod" not found`),
			fake.NewSimpleClientset(&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "other-namespace",
				},
			}),
			"test-namespace",
			"test-pod",
		},
		{
			"Get 1 pod in same correct namespace",
			&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			},
			nil,
			fake.NewSimpleClientset(&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			}),
			"test-namespace",
			"test-pod",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := &podclient.OnDemandPodNamespaceLister{
				Namespace: test.namespace,
				Clientset: test.clientset,
			}
			pods, err := gatherer.Get(test.name)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, pods) {
				t.Errorf("pods mismatch (-want +got):\n%s", cmp.Diff(test.expected, pods))
			}
		})
	}
}
