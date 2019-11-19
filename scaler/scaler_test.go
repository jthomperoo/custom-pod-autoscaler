/*
Copyright 2019 The Custom Pod Autoscaler Authors.

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

package scaler_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/scaler"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	fakeappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1/fake"
	k8stesting "k8s.io/client-go/testing"
)

type getMetricer struct {
	getMetrics func(deployment *appsv1.Deployment) (*metric.ResourceMetrics, error)
}

func (m *getMetricer) GetMetrics(deployment *appsv1.Deployment) (*metric.ResourceMetrics, error) {
	return m.getMetrics(deployment)
}

type getEvaluationer struct {
	getEvaluation func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error)
}

func (e *getEvaluationer) GetEvaluation(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
	return e.getEvaluation(resourceMetrics)
}

func TestScaler(t *testing.T) {

	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description     string
		expected        error
		clientset       kubernetes.Interface
		config          *config.Config
		getMetricer     *getMetricer
		getEvaluationer *getEvaluationer
	}{
		{
			"Deployment not found",
			errors.New(`deployments.apps "test" not found`),
			fake.NewSimpleClientset(),
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			nil,
			nil,
		},
		{
			"Gather metric fail",
			errors.New("fail to get metric"),
			fake.NewSimpleClientset(&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test namespace",
				},
			}),
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			func() *getMetricer {
				getMetric := getMetricer{}
				getMetric.getMetrics = func(deployment *appsv1.Deployment) (*metric.ResourceMetrics, error) {
					return nil, errors.New("fail to get metric")
				}
				return &getMetric
			}(),
			nil,
		},
		{
			"Evaluate fail",
			errors.New("fail to evaluate"),
			fake.NewSimpleClientset(&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test namespace",
				},
			}),
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			func() *getMetricer {
				getMetric := getMetricer{}
				getMetric.getMetrics = func(deployment *appsv1.Deployment) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *getEvaluationer {
				getEvaluation := getEvaluationer{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return nil, errors.New("fail to evaluate")
				}
				return &getEvaluation
			}(),
		},
		{
			"Fail to update deployment",
			errors.New("fail to update deployment"),
			func() *fake.Clientset {
				replicas := int32(3)
				clientset := fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &replicas,
					},
				})
				clientset.AppsV1().(*fakeappsv1.FakeAppsV1).Fake.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("fail to update deployment")
				})
				return clientset
			}(),
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			func() *getMetricer {
				getMetric := getMetricer{}
				getMetric.getMetrics = func(deployment *appsv1.Deployment) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *getEvaluationer {
				getEvaluation := getEvaluationer{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(1),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Success, autoscaling disabled",
			nil,
			func() *fake.Clientset {
				replicas := int32(0)
				clientset := fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &replicas,
					},
				})
				return clientset
			}(),
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			nil,
			nil,
		},
		{
			"Success, no change in scale",
			nil,
			func() *fake.Clientset {
				replicas := int32(3)
				clientset := fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &replicas,
					},
				})
				return clientset
			}(),
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			func() *getMetricer {
				getMetric := getMetricer{}
				getMetric.getMetrics = func(deployment *appsv1.Deployment) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *getEvaluationer {
				getEvaluation := getEvaluationer{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(3),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Success evaluation above max replicas, scale to max replicas",
			nil,
			func() *fake.Clientset {
				replicas := int32(3)
				clientset := fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &replicas,
					},
				})
				clientset.AppsV1().(*fakeappsv1.FakeAppsV1).Fake.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
					update, ok := action.(k8stesting.UpdateAction)
					if !ok {
						return true, nil, errors.New("fail to cast action to update action")
					}
					obj := update.GetObject()
					deployment := reflect.ValueOf(obj).Elem()
					spec, ok := deployment.FieldByName("Spec").Interface().(appsv1.DeploymentSpec)
					if !ok {
						return true, nil, errors.New("fail to cast runtime.object to deployment")
					}
					replicas := int32(5)
					if *spec.Replicas != 3 {
						return true, nil, fmt.Errorf("Replicas mismatch (-want +got):\n%s", cmp.Diff(replicas, spec.Replicas))
					}
					return true, nil, nil
				})
				return clientset
			}(),
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				MinReplicas: 1,
				MaxReplicas: 3,
			},
			func() *getMetricer {
				getMetric := getMetricer{}
				getMetric.getMetrics = func(deployment *appsv1.Deployment) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *getEvaluationer {
				getEvaluation := getEvaluationer{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(5),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Success evaluation below min replicas, scale to min replicas",
			nil,
			func() *fake.Clientset {
				replicas := int32(3)
				clientset := fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &replicas,
					},
				})
				clientset.AppsV1().(*fakeappsv1.FakeAppsV1).Fake.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
					update, ok := action.(k8stesting.UpdateAction)
					if !ok {
						return true, nil, errors.New("fail to cast action to update action")
					}
					obj := update.GetObject()
					deployment := reflect.ValueOf(obj).Elem()
					spec, ok := deployment.FieldByName("Spec").Interface().(appsv1.DeploymentSpec)
					if !ok {
						return true, nil, errors.New("fail to cast runtime.object to deployment")
					}
					replicas := int32(2)
					if *spec.Replicas != replicas {
						return true, nil, fmt.Errorf("Replicas mismatch (-want +got):\n%s", cmp.Diff(replicas, spec.Replicas))
					}
					return true, nil, nil
				})
				return clientset
			}(),
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				MinReplicas: 2,
				MaxReplicas: 10,
			},
			func() *getMetricer {
				getMetric := getMetricer{}
				getMetric.getMetrics = func(deployment *appsv1.Deployment) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *getEvaluationer {
				getEvaluation := getEvaluationer{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(1),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Success evaluation within min-max bounds, scale to evaluation",
			nil,
			func() *fake.Clientset {
				replicas := int32(3)
				clientset := fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test namespace",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &replicas,
					},
				})
				clientset.AppsV1().(*fakeappsv1.FakeAppsV1).Fake.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
					update, ok := action.(k8stesting.UpdateAction)
					if !ok {
						return true, nil, errors.New("fail to cast action to update action")
					}
					obj := update.GetObject()
					deployment := reflect.ValueOf(obj).Elem()
					spec, ok := deployment.FieldByName("Spec").Interface().(appsv1.DeploymentSpec)
					if !ok {
						return true, nil, errors.New("fail to cast runtime.object to deployment")
					}
					replicas := int32(7)
					if *spec.Replicas != replicas {
						return true, nil, fmt.Errorf("Replicas mismatch (-want +got):\n%s", cmp.Diff(replicas, spec.Replicas))
					}
					return true, nil, nil
				})
				return clientset
			}(),
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				MinReplicas: 1,
				MaxReplicas: 10,
			},
			func() *getMetricer {
				getMetric := getMetricer{}
				getMetric.getMetrics = func(deployment *appsv1.Deployment) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *getEvaluationer {
				getEvaluation := getEvaluationer{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(7),
					}, nil
				}
				return &getEvaluation
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			deploymentInterface := test.clientset.AppsV1().Deployments(test.config.Namespace)
			scaler := &scaler.Scaler{
				Clientset:         test.clientset,
				DeploymentsClient: deploymentInterface,
				Config:            test.config,
				GetMetricer:       test.getMetricer,
				GetEvaluationer:   test.getEvaluationer,
			}

			err := scaler.Scale()
			if !cmp.Equal(&err, &test.expected, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expected, err, equateErrorMessage))
			}
		})
	}
}
