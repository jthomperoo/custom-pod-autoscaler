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
// +build unit

package metricget_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/autoscaler"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/metricget"
	"github.com/jthomperoo/custom-pod-autoscaler/k8smetric"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	fakeappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestGetMetrics(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expectedErr error
		expected    []*metric.ResourceMetric
		spec        metric.Info
		gatherer    metricget.Gatherer
	}{
		{
			"Invalid run mode",
			errors.New("Unknown run mode: invalid"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   "invalid",
				},
				Clientset: func() *k8sfake.Clientset {
					clientset := k8sfake.NewSimpleClientset()
					clientset.AppsV1().(*fakeappsv1.FakeAppsV1).Fake.PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, nil, errors.New("fail to list pods")
					})
					return clientset
				}(),
			},
		},
		{
			"Per pod unsupported resource selector",
			errors.New("Unsupported resource of type *v1.DaemonSet"),
			nil,
			metric.Info{
				Resource: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
			},
		},
		{
			"Per pod fail to get deployment selector",
			errors.New(`"invalid" is not a valid selector operator`),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Operator: "invalid",
								},
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
			},
		},
		{
			"Per pod fail to get replicaset selector",
			errors.New(`"invalid" is not a valid selector operator`),
			nil,
			metric.Info{
				Resource: &appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.ReplicaSetSpec{
						Selector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Operator: "invalid",
								},
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
			},
		},
		{
			"Per pod fail to get statefulset selector",
			errors.New(`"invalid" is not a valid selector operator`),
			nil,
			metric.Info{
				Resource: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.StatefulSetSpec{
						Selector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Operator: "invalid",
								},
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
			},
		},
		{
			"Per pod error when listing pods",
			errors.New("fail to list pods"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: func() *k8sfake.Clientset {
					clientset := k8sfake.NewSimpleClientset()
					clientset.AppsV1().(*fakeappsv1.FakeAppsV1).Fake.PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, nil, errors.New("fail to list pods")
					})
					return clientset
				}(),
			},
		},
		{
			"Per pod pre-metric hook fail",
			errors.New("pre-metric hook fail"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test",
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
					PreMetric: &config.Method{
						Type: "fake",
					},
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "", errors.New("pre-metric hook fail")
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod single pod single deployment shell execute fail",
			errors.New("fail to get metric"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test",
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "", errors.New("fail to get metric")
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod no resources",
			nil,
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{},
				RunType:  autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod post-metric hook fail",
			errors.New("post-metric hook fail"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test",
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
					Metric: &config.Method{
						Type: "metric",
					},
					PostMetric: &config.Method{
						Type: "post-metric",
					},
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						if method.Type == "post-metric" {
							return "", errors.New("post-metric hook fail")
						}
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod no pod in managed deployment, but pod in other deployment with different name in same namespace",
			nil,
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test managed deployment",
						Namespace: "test managed namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test-managed",
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test managed namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test managed namespace",
							Labels:    map[string]string{"app": "test-unmanaged"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod no pod in managed deployment, but pod in other deployment with same name in different namespace",
			nil,
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test managed deployment",
						Namespace: "test managed namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test-managed",
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test managed namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test unmanaged namespace",
							Labels:    map[string]string{"app": "test-managed"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod single pod single deployment shell execute success",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test",
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod single pod single deployment shell execute success with pre-metric hook",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test",
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
					Metric: &config.Method{
						Type: "metric",
					},
					PreMetric: &config.Method{
						Type: "pre-metric",
					},
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod single pod single deployment shell execute success with post-metric hook",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test",
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
					Metric: &config.Method{
						Type: "metric",
					},
					PostMetric: &config.Method{
						Type: "post-metric",
					},
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod single pod, single replicaset success",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.ReplicaSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test",
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod single pod, single statefulset success",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.StatefulSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test",
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod single pod, single replicationcontroller success",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &corev1.ReplicationController{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: corev1.ReplicationControllerSpec{
						Selector: map[string]string{
							"app": "test",
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per pod multiple pod single deployment shell execute success",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "first pod",
					Value:    "test value",
				},
				{
					Resource: "second pod",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test",
							},
						},
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "first pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "second pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
				),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per resource shell execute fail",
			errors.New("fail to get metric"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerResourceRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "", errors.New("fail to get metric")
					}
					return &execute
				}(),
			},
		},
		{
			"Per resource pre-metric hook fail",
			errors.New("pre-metric hook fail"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerResourceRunMode,
					PreMetric: &config.Method{
						Type: "pre-metric",
					},
				},
				Clientset: k8sfake.NewSimpleClientset(),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "", errors.New("pre-metric hook fail")
					}
					return &execute
				}(),
			},
		},
		{
			"Per resource post-metric hook fail",
			errors.New("post-metric hook fail"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerResourceRunMode,
					Metric: &config.Method{
						Type: "metric",
					},
					PostMetric: &config.Method{
						Type: "post-metric",
					},
				},
				Clientset: k8sfake.NewSimpleClientset(),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						if method.Type == "post-metric" {
							return "", errors.New("post-metric hook fail")
						}
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per resource shell execute success",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test deployment",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerResourceRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per resource shell execute success with pre-metric hook",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test deployment",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerResourceRunMode,
					PreMetric: &config.Method{
						Type: "pre-metric",
					},
				},
				Clientset: k8sfake.NewSimpleClientset(),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per resource shell execute success with post-metric hook",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test deployment",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerResourceRunMode,
					PostMetric: &config.Method{
						Type: "pre-metric",
					},
				},
				Clientset: k8sfake.NewSimpleClientset(),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
			},
		},
		{
			"Per resource shell execute success with K8s metrics",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test deployment",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerResourceRunMode,
					PostMetric: &config.Method{
						Type: "pre-metric",
					},
					KubernetesMetricSpecs: []config.K8sMetricSpec{
						{
							Type: v2beta2.ResourceMetricSourceType,
							Resource: &config.K8sResourceMetricSource{
								Name: v1.ResourceCPU,
								Target: config.K8sMetricTarget{
									Type: v2beta2.AverageValueMetricType,
								},
							},
						},
					},
				},
				Clientset: k8sfake.NewSimpleClientset(),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
				K8sMetricGatherer: &fake.Gather{
					GetMetricsReactor: func(resource metav1.Object, specs []config.K8sMetricSpec, namespace string) ([]*k8smetric.Metric, error) {
						return []*k8smetric.Metric{
							{
								CurrentReplicas: 3,
							},
						}, nil
					},
				},
			},
		},
		{
			"Per resource shell execute success, fail to get K8s metrics, but RequireKubernetesMetrics: false",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test deployment",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerResourceRunMode,
					PostMetric: &config.Method{
						Type: "pre-metric",
					},
					RequireKubernetesMetrics: false,
					KubernetesMetricSpecs: []config.K8sMetricSpec{
						{
							Type: v2beta2.ResourceMetricSourceType,
							Resource: &config.K8sResourceMetricSource{
								Name: v1.ResourceCPU,
								Target: config.K8sMetricTarget{
									Type: v2beta2.AverageValueMetricType,
								},
							},
						},
					},
				},
				Clientset: k8sfake.NewSimpleClientset(),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
				K8sMetricGatherer: &fake.Gather{
					GetMetricsReactor: func(resource metav1.Object, specs []config.K8sMetricSpec, namespace string) ([]*k8smetric.Metric, error) {
						return nil, errors.New("fail to get K8s metrics!")
					},
				},
			},
		},
		{
			"Per resource shell execute failure, fail to get K8s metrics, RequireKubernetesMetrics: true",
			errors.New("fail to get K8s metrics!"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerResourceRunMode,
					PostMetric: &config.Method{
						Type: "pre-metric",
					},
					RequireKubernetesMetrics: true,
					KubernetesMetricSpecs: []config.K8sMetricSpec{
						{
							Type: v2beta2.ResourceMetricSourceType,
							Resource: &config.K8sResourceMetricSource{
								Name: v1.ResourceCPU,
								Target: config.K8sMetricTarget{
									Type: v2beta2.AverageValueMetricType,
								},
							},
						},
					},
				},
				Clientset: k8sfake.NewSimpleClientset(),
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
						return "test value", nil
					}
					return &execute
				}(),
				K8sMetricGatherer: &fake.Gather{
					GetMetricsReactor: func(resource metav1.Object, specs []config.K8sMetricSpec, namespace string) ([]*k8smetric.Metric, error) {
						return nil, errors.New("fail to get K8s metrics!")
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			metrics, err := test.gatherer.GetMetrics(test.spec)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if test.expectedErr != nil {
				return
			}
			if !cmp.Equal(metrics, test.expected) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, metrics))
			}
		})
	}
}
