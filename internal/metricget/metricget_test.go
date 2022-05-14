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

package metricget_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/metricget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/metric"
	"github.com/jthomperoo/k8shorizmetrics/metrics"
	"github.com/jthomperoo/k8shorizmetrics/metrics/podmetrics"
	"github.com/jthomperoo/k8shorizmetrics/metrics/resource"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	fakeappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1/fake"
	k8stesting "k8s.io/client-go/testing"
)

func selectorFromString(selectorStr string) labels.Selector {
	selector, _ := labels.Parse(selectorStr)
	return selector
}

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
		podSelector labels.Selector
	}{
		{
			"Invalid run mode",
			errors.New("unknown run mode: invalid"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: config.ScalerRunType,
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
			nil,
		},
		{
			"Per pod error when listing pods",
			errors.New("failed to get pods being managed: fail to list pods"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: config.ScalerRunType,
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
			labels.NewSelector(),
		},
		{
			"Per pod pre-metric hook fail",
			errors.New("failed to run pre-metric hook: pre-metric hook fail"),
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
				RunType: config.ScalerRunType,
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
					&corev1.Pod{
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
			labels.NewSelector(),
		},
		{
			"Per pod single pod single deployment shell execute fail",
			errors.New("failed to gather metrics: fail to get metric"),
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
				RunType: config.ScalerRunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&corev1.Pod{
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
			labels.NewSelector(),
		},
		{
			"Per pod no resources",
			nil,
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{},
				RunType:  config.ScalerRunType,
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
			labels.NewSelector(),
		},
		{
			"Per pod post-metric hook fail",
			errors.New("failed to run post-metric hook: post-metric hook fail"),
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
				RunType: config.ScalerRunType,
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
					&corev1.Pod{
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
			labels.NewSelector(),
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
				RunType: config.ScalerRunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test managed namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&corev1.Pod{
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
			selectorFromString("app==test"),
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
				RunType: config.ScalerRunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test managed namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&corev1.Pod{
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
			labels.NewSelector(),
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
				RunType: config.ScalerRunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&corev1.Pod{
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
			labels.NewSelector(),
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
				RunType: config.ScalerRunType,
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
					&corev1.Pod{
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
			labels.NewSelector(),
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
				RunType: config.ScalerRunType,
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
					&corev1.Pod{
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
			labels.NewSelector(),
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
				RunType: config.ScalerRunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&corev1.Pod{
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
			labels.NewSelector(),
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
				RunType: config.ScalerRunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&corev1.Pod{
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
			labels.NewSelector(),
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
				RunType: config.ScalerRunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&corev1.Pod{
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
			labels.NewSelector(),
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
				RunType: config.ScalerRunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "first pod",
							Namespace: "test namespace",
							Labels:    map[string]string{"app": "test"},
						},
					},
					&corev1.Pod{
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
			labels.NewSelector(),
		},
		{
			"Per resource shell execute fail",
			errors.New("failed to gather metrics: fail to get metric"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: config.ScalerRunType,
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
			nil,
		},
		{
			"Per resource pre-metric hook fail",
			errors.New("failed to run pre-metric hook: pre-metric hook fail"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: config.ScalerRunType,
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
			nil,
		},
		{
			"Per resource post-metric hook fail",
			errors.New("failed to run post-metric hook: post-metric hook fail"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: config.ScalerRunType,
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
			nil,
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
				RunType: config.ScalerRunType,
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
			nil,
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
				RunType: config.ScalerRunType,
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
			nil,
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
				RunType: config.ScalerRunType,
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
			nil,
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
				RunType: config.ScalerRunType,
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
							Type: autoscalingv2.ResourceMetricSourceType,
							Resource: &autoscalingv2.ResourceMetricSource{
								Name: corev1.ResourceCPU,
								Target: autoscalingv2.MetricTarget{
									Type: autoscalingv2.AverageValueMetricType,
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
				K8sMetricGatherer: &fake.Gatherer{
					GatherReactor: func(specs []autoscalingv2.MetricSpec, namespace string, podSelector labels.Selector) ([]*metrics.Metric, error) {
						return []*metrics.Metric{
							{
								Resource: &resource.Metric{
									PodMetricsInfo: podmetrics.MetricsInfo{
										"test": podmetrics.Metric{
											Value: 5,
										},
									},
								},
							},
						}, nil
					},
				},
			},
			labels.NewSelector(),
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
				RunType: config.ScalerRunType,
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
							Type: autoscalingv2.ResourceMetricSourceType,
							Resource: &autoscalingv2.ResourceMetricSource{
								Name: corev1.ResourceCPU,
								Target: autoscalingv2.MetricTarget{
									Type: autoscalingv2.AverageValueMetricType,
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
				K8sMetricGatherer: &fake.Gatherer{
					GatherReactor: func(specs []autoscalingv2.MetricSpec, namespace string, podSelector labels.Selector) ([]*metrics.Metric, error) {
						return nil, errors.New("fail to get K8s metrics!")
					},
				},
			},
			labels.NewSelector(),
		},
		{
			"Per resource shell execute failure, fail to get K8s metrics, RequireKubernetesMetrics: true",
			errors.New("failed to get required Kubernetes metrics: fail to get K8s metrics!"),
			nil,
			metric.Info{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: config.ScalerRunType,
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
							Type: autoscalingv2.ResourceMetricSourceType,
							Resource: &autoscalingv2.ResourceMetricSource{
								Name: corev1.ResourceCPU,
								Target: autoscalingv2.MetricTarget{
									Type: autoscalingv2.AverageValueMetricType,
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
				K8sMetricGatherer: &fake.Gatherer{
					GatherReactor: func(specs []autoscalingv2.MetricSpec, namespace string, podSelector labels.Selector) ([]*metrics.Metric, error) {
						return nil, errors.New("fail to get K8s metrics!")
					},
				},
			},
			labels.NewSelector(),
		},
		{
			"Per pod single pod, single Argo Rollout success",
			nil,
			[]*metric.ResourceMetric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Info{
				Resource: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"name":      "test deployment",
						"namespace": "test namespace",
						"spec": map[string]interface{}{
							"selector": map[string]interface{}{
								"matchLabels": map[string]string{
									"test": "test",
								},
							},
						},
					},
				},
				RunType: config.ScalerRunType,
			},
			metricget.Gatherer{
				Config: &config.Config{
					Namespace: "test namespace",
					RunMode:   config.PerPodRunMode,
				},
				Clientset: k8sfake.NewSimpleClientset(
					&corev1.Pod{
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
			labels.NewSelector(),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			metrics, err := test.gatherer.GetMetrics(test.spec, test.podSelector)
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
