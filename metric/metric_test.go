/*
Copyright 2020 The Custom Pod Autoscaler Authors.

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

package metric_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/autoscaler"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/execute"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
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
		expected    []*metric.Metric
		spec        metric.Spec
		config      *config.Config
		clientset   kubernetes.Interface
		execute     execute.Executer
	}{
		{
			"Invalid run mode",
			errors.New("Unknown run mode: invalid"),
			nil,
			metric.Spec{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			&config.Config{
				Namespace: "test namespace",
				RunMode:   "invalid",
			},
			func() *k8sfake.Clientset {
				clientset := k8sfake.NewSimpleClientset()
				clientset.AppsV1().(*fakeappsv1.FakeAppsV1).Fake.PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("fail to list pods")
				})
				return clientset
			}(),
			nil,
		},
		{
			"Per pod unsupported resource selector",
			errors.New("Unsupported resource of type *v1.DaemonSet"),
			nil,
			metric.Spec{
				Resource: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			nil,
			nil,
		},
		{
			"Per pod fail to get deployment selector",
			errors.New(`"invalid" is not a valid selector operator`),
			nil,
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			nil,
			nil,
		},
		{
			"Per pod fail to get replicaset selector",
			errors.New(`"invalid" is not a valid selector operator`),
			nil,
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			nil,
			nil,
		},
		{
			"Per pod fail to get statefulset selector",
			errors.New(`"invalid" is not a valid selector operator`),
			nil,
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			nil,
			nil,
		},
		{
			"Per pod error when listing pods",
			errors.New("fail to list pods"),
			nil,
			metric.Spec{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			func() *k8sfake.Clientset {
				clientset := k8sfake.NewSimpleClientset()
				clientset.AppsV1().(*fakeappsv1.FakeAppsV1).Fake.PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("fail to list pods")
				})
				return clientset
			}(),
			nil,
		},
		{
			"Per pod pre-metric hook fail",
			errors.New("pre-metric hook fail"),
			nil,
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
				PreMetric: &config.Method{
					Type: "fake",
				},
			},
			k8sfake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test pod",
						Namespace: "test namespace",
						Labels:    map[string]string{"app": "test"},
					},
				},
			),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "", errors.New("pre-metric hook fail")
				}
				return &execute
			}(),
		},
		{
			"Per pod single pod single deployment shell execute fail",
			errors.New("fail to get metric"),
			nil,
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			k8sfake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test pod",
						Namespace: "test namespace",
						Labels:    map[string]string{"app": "test"},
					},
				},
			),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "", errors.New("fail to get metric")
				}
				return &execute
			}(),
		},
		{
			"Per pod no resources",
			nil,
			nil,
			metric.Spec{
				Resource: &appsv1.Deployment{},
				RunType:  autoscaler.RunType,
			},
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			k8sfake.NewSimpleClientset(),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per pod post-metric hook fail",
			errors.New("post-metric hook fail"),
			nil,
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
				Metric: &config.Method{
					Type: "metric",
				},
				PostMetric: &config.Method{
					Type: "post-metric",
				},
			},
			k8sfake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test pod",
						Namespace: "test namespace",
						Labels:    map[string]string{"app": "test"},
					},
				},
			),
			func() *fake.Execute {
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
		{
			"Per pod no pod in managed deployment, but pod in other deployment with different name in same namespace",
			nil,
			nil,
			metric.Spec{
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
			&config.Config{
				Namespace: "test managed namespace",
				RunMode:   config.PerPodRunMode,
			},
			k8sfake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test pod",
						Namespace: "test managed namespace",
						Labels:    map[string]string{"app": "test-unmanaged"},
					},
				},
			),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per pod no pod in managed deployment, but pod in other deployment with same name in different namespace",
			nil,
			nil,
			metric.Spec{
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
			&config.Config{
				Namespace: "test managed namespace",
				RunMode:   config.PerPodRunMode,
			},
			k8sfake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test pod",
						Namespace: "test unmanaged namespace",
						Labels:    map[string]string{"app": "test-managed"},
					},
				},
			),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per pod single pod single deployment shell execute success",
			nil,
			[]*metric.Metric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			k8sfake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test pod",
						Namespace: "test namespace",
						Labels:    map[string]string{"app": "test"},
					},
				},
			),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per pod single pod single deployment shell execute success with pre-metric hook",
			nil,
			[]*metric.Metric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
				Metric: &config.Method{
					Type: "metric",
				},
				PreMetric: &config.Method{
					Type: "pre-metric",
				},
			},
			k8sfake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test pod",
						Namespace: "test namespace",
						Labels:    map[string]string{"app": "test"},
					},
				},
			),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per pod single pod single deployment shell execute success with post-metric hook",
			nil,
			[]*metric.Metric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
				Metric: &config.Method{
					Type: "metric",
				},
				PostMetric: &config.Method{
					Type: "post-metric",
				},
			},
			k8sfake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test pod",
						Namespace: "test namespace",
						Labels:    map[string]string{"app": "test"},
					},
				},
			),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per pod single pod, single replicaset success",
			nil,
			[]*metric.Metric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			k8sfake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test pod",
						Namespace: "test namespace",
						Labels:    map[string]string{"app": "test"},
					},
				},
			),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per pod single pod, single statefulset success",
			nil,
			[]*metric.Metric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			k8sfake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test pod",
						Namespace: "test namespace",
						Labels:    map[string]string{"app": "test"},
					},
				},
			),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per pod single pod, single replicationcontroller success",
			nil,
			[]*metric.Metric{
				{
					Resource: "test pod",
					Value:    "test value",
				},
			},
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			k8sfake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test pod",
						Namespace: "test namespace",
						Labels:    map[string]string{"app": "test"},
					},
				},
			),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per pod multiple pod single deployment shell execute success",
			nil,
			[]*metric.Metric{
				{
					Resource: "first pod",
					Value:    "test value",
				},
				{
					Resource: "second pod",
					Value:    "test value",
				},
			},
			metric.Spec{
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
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerPodRunMode,
			},
			k8sfake.NewSimpleClientset(
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
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per resource shell execute fail",
			errors.New("fail to get metric"),
			nil,
			metric.Spec{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerResourceRunMode,
			},
			k8sfake.NewSimpleClientset(),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "", errors.New("fail to get metric")
				}
				return &execute
			}(),
		},
		{
			"Per resource pre-metric hook fail",
			errors.New("pre-metric hook fail"),
			nil,
			metric.Spec{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerResourceRunMode,
				PreMetric: &config.Method{
					Type: "pre-metric",
				},
			},
			k8sfake.NewSimpleClientset(),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "", errors.New("pre-metric hook fail")
				}
				return &execute
			}(),
		},
		{
			"Per resource post-metric hook fail",
			errors.New("post-metric hook fail"),
			nil,
			metric.Spec{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerResourceRunMode,
				Metric: &config.Method{
					Type: "metric",
				},
				PostMetric: &config.Method{
					Type: "post-metric",
				},
			},
			k8sfake.NewSimpleClientset(),
			func() *fake.Execute {
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
		{
			"Per resource shell execute success",
			nil,
			[]*metric.Metric{
				{
					Resource: "test deployment",
					Value:    "test value",
				},
			},
			metric.Spec{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerResourceRunMode,
			},
			k8sfake.NewSimpleClientset(),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per resource shell execute success with pre-metric hook",
			nil,
			[]*metric.Metric{
				{
					Resource: "test deployment",
					Value:    "test value",
				},
			},
			metric.Spec{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerResourceRunMode,
				PreMetric: &config.Method{
					Type: "pre-metric",
				},
			},
			k8sfake.NewSimpleClientset(),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
		{
			"Per resource shell execute success with post-metric hook",
			nil,
			[]*metric.Metric{
				{
					Resource: "test deployment",
					Value:    "test value",
				},
			},
			metric.Spec{
				Resource: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test deployment",
						Namespace: "test namespace",
					},
				},
				RunType: autoscaler.RunType,
			},
			&config.Config{
				Namespace: "test namespace",
				RunMode:   config.PerResourceRunMode,
				PostMetric: &config.Method{
					Type: "pre-metric",
				},
			},
			k8sfake.NewSimpleClientset(),
			func() *fake.Execute {
				execute := fake.Execute{}
				execute.ExecuteWithValueReactor = func(method *config.Method, value string) (string, error) {
					return "test value", nil
				}
				return &execute
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := &metric.Gatherer{
				Clientset: test.clientset,
				Config:    test.config,
				Execute:   test.execute,
			}
			metrics, err := gatherer.GetMetrics(test.spec)
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
