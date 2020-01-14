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
// +build unit

package scaler_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/resourceclient"
	"github.com/jthomperoo/custom-pod-autoscaler/scaler"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/scale"
	scaleFake "k8s.io/client-go/scale/fake"
	k8stesting "k8s.io/client-go/testing"
)

type fakeGetMetric struct {
	getMetrics func(resource metav1.Object) (*metric.ResourceMetrics, error)
}

func (m *fakeGetMetric) GetMetrics(resource metav1.Object) (*metric.ResourceMetrics, error) {
	return m.getMetrics(resource)
}

type fakeGetEvaluation struct {
	getEvaluation func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error)
}

func (e *fakeGetEvaluation) GetEvaluation(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
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
		client          resourceclient.Client
		scaler          scale.ScalesGetter
		config          *config.Config
		getMetricer     metric.GetMetricer
		getEvaluationer evaluate.GetEvaluationer
	}{
		{
			"Get resource fail",
			errors.New(`fail to get resource`),
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return nil, errors.New("fail to get resource")
				},
			},
			nil,
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
			"Unsupported resource",
			errors.New(`Unsupported resource of type *v1.DaemonSet`),
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.DaemonSet{}, nil
				},
			},
			nil,
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "daemonset",
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
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: namespace,
						},
					}, nil
				},
			},
			nil,
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return nil, errors.New("fail to get metric")
				}
				return &getMetric
			}(),
			nil,
		},
		{
			"Evaluate fail",
			errors.New("fail to evaluate"),
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: namespace,
						},
					}, nil
				},
			},
			nil,
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return nil, errors.New("fail to evaluate")
				}
				return &getEvaluation
			}(),
		},
		{
			"Fail to parse group version",
			errors.New("unexpected GroupVersion string: /invalid/"),
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(3)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}, nil
				},
			},
			nil,
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "/invalid/",
				},
			},
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(1),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Fail to get scale for resource",
			errors.New("fail to get resource"),
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(3)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}, nil
				},
			},
			&scaleFake.FakeScaleClient{
				Fake: k8stesting.Fake{
					ReactionChain: []k8stesting.Reactor{
						&k8stesting.SimpleReactor{
							Resource: "deployment",
							Verb:     "get",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, nil, errors.New("fail to get resource")
							},
						},
					},
				},
			},
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(1),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Fail to update scale for resource",
			errors.New("fail to update resource"),
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(3)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}, nil
				},
			},
			&scaleFake.FakeScaleClient{
				Fake: k8stesting.Fake{
					ReactionChain: []k8stesting.Reactor{
						&k8stesting.SimpleReactor{
							Resource: "deployment",
							Verb:     "get",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{
									Spec: autoscaling.ScaleSpec{
										Replicas: 3,
									},
								}, nil
							},
						},
						&k8stesting.SimpleReactor{
							Resource: "deployment",
							Verb:     "update",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, nil, errors.New("fail to update resource")
							},
						},
					},
				},
			},
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(1),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Success, deployment, autoscaling disabled",
			nil,
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(0)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}, nil
				},
			},
			nil,
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
			"Success, deployment, no change in scale",
			nil,
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(3)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}, nil
				},
			},
			nil,
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
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(3),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Success, deployment, evaluation above max replicas, scale to max replicas",
			nil,
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(2)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}, nil
				},
			},
			&scaleFake.FakeScaleClient{
				Fake: k8stesting.Fake{
					ReactionChain: []k8stesting.Reactor{
						&k8stesting.SimpleReactor{
							Resource: "deployment",
							Verb:     "get",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{
									Spec: autoscaling.ScaleSpec{
										Replicas: 3,
									},
								}, nil
							},
						},
						&k8stesting.SimpleReactor{
							Resource: "deployment",
							Verb:     "update",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{}, nil
							},
						},
					},
				},
			},
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
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(5),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Success, deployment, evaluation below min replicas, scale to min replicas",
			nil,
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(5)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}, nil
				},
			},
			&scaleFake.FakeScaleClient{
				Fake: k8stesting.Fake{
					ReactionChain: []k8stesting.Reactor{
						&k8stesting.SimpleReactor{
							Resource: "deployment",
							Verb:     "get",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{
									Spec: autoscaling.ScaleSpec{
										Replicas: 2,
									},
								}, nil
							},
						},
						&k8stesting.SimpleReactor{
							Resource: "deployment",
							Verb:     "update",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{}, nil
							},
						},
					},
				},
			},
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
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(1),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Success, deployment, evaluation within min-max bounds, scale to evaluation",
			nil,
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(5)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}, nil
				},
			},
			&scaleFake.FakeScaleClient{
				Fake: k8stesting.Fake{
					ReactionChain: []k8stesting.Reactor{
						&k8stesting.SimpleReactor{
							Resource: "deployment",
							Verb:     "get",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{
									Spec: autoscaling.ScaleSpec{
										Replicas: 7,
									},
								}, nil
							},
						},
						&k8stesting.SimpleReactor{
							Resource: "deployment",
							Verb:     "update",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{}, nil
							},
						},
					},
				},
			},
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
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(7),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Success, replicaset, evaluation within min-max bounds, scale to evaluation",
			nil,
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(3)
					return &appsv1.ReplicaSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.ReplicaSetSpec{
							Replicas: &replicas,
						},
					}, nil
				},
			},
			&scaleFake.FakeScaleClient{
				Fake: k8stesting.Fake{
					ReactionChain: []k8stesting.Reactor{
						&k8stesting.SimpleReactor{
							Resource: "replicaset",
							Verb:     "get",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{
									Spec: autoscaling.ScaleSpec{
										Replicas: 4,
									},
								}, nil
							},
						},
						&k8stesting.SimpleReactor{
							Resource: "replicaset",
							Verb:     "update",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{}, nil
							},
						},
					},
				},
			},
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "replicaset",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				MinReplicas: 1,
				MaxReplicas: 10,
			},
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(7),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Success, replicationcontroller, evaluation within min-max bounds, scale to evaluation",
			nil,
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(8)
					return &corev1.ReplicationController{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: corev1.ReplicationControllerSpec{
							Replicas: &replicas,
						},
					}, nil
				},
			},
			&scaleFake.FakeScaleClient{
				Fake: k8stesting.Fake{
					ReactionChain: []k8stesting.Reactor{
						&k8stesting.SimpleReactor{
							Resource: "replicationcontroller",
							Verb:     "get",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{
									Spec: autoscaling.ScaleSpec{
										Replicas: 2,
									},
								}, nil
							},
						},
						&k8stesting.SimpleReactor{
							Resource: "replicationcontroller",
							Verb:     "update",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{}, nil
							},
						},
					},
				},
			},
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "replicationcontroller",
					Name:       "test",
					APIVersion: "v1",
				},
				MinReplicas: 1,
				MaxReplicas: 10,
			},
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: int32(7),
					}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Success, statefulset, evaluation within min-max bounds, scale to evaluation",
			nil,
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(1)
					return &appsv1.StatefulSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.StatefulSetSpec{
							Replicas: &replicas,
						},
					}, nil
				},
			},
			&scaleFake.FakeScaleClient{
				Fake: k8stesting.Fake{
					ReactionChain: []k8stesting.Reactor{
						&k8stesting.SimpleReactor{
							Resource: "statefulset",
							Verb:     "get",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{
									Spec: autoscaling.ScaleSpec{
										Replicas: 10,
									},
								}, nil
							},
						},
						&k8stesting.SimpleReactor{
							Resource: "statefulset",
							Verb:     "update",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{}, nil
							},
						},
					},
				},
			},
			&config.Config{
				Namespace: "test namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "statefulset",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				MinReplicas: 1,
				MaxReplicas: 10,
			},
			func() *fakeGetMetric {
				getMetric := fakeGetMetric{}
				getMetric.getMetrics = func(resource metav1.Object) (*metric.ResourceMetrics, error) {
					return &metric.ResourceMetrics{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
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
			scaler := &scaler.Scaler{
				Scaler:          test.scaler,
				Client:          test.client,
				Config:          test.config,
				GetMetricer:     test.getMetricer,
				GetEvaluationer: test.getEvaluationer,
			}

			err := scaler.Scale()
			if !cmp.Equal(&err, &test.expected, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expected, err, equateErrorMessage))
			}
		})
	}
}
