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

package scale_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/autoscaler"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/execute"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/scale"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sscale "k8s.io/client-go/scale"
	scaleFake "k8s.io/client-go/scale/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestScale_Scale(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    *evaluate.Evaluation
		expectedErr error
		scaler      k8sscale.ScalesGetter
		config      *config.Config
		executer    execute.Executer
		spec        scale.Spec
	}{
		{
			"Unsupported resource",
			nil,
			errors.New(`Unsupported resource of type *v1.DaemonSet`),
			nil,
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation:  evaluate.Evaluation{},
				Resource:    &appsv1.DaemonSet{},
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "daemonset",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Fail to parse group version",
			nil,
			errors.New("unexpected GroupVersion string: /invalid/"),
			nil,
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "deployment",
						APIVersion: "/invalid/",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "/invalid/",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Fail to get scale for resource",
			nil,
			errors.New("fail to get resource"),
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
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Fail to update scale for resource",
			nil,
			errors.New("fail to update resource"),
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
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(1),
				},
				Resource: func() *appsv1.Deployment {
					replicas := int32(3)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Fail to run pre-scaling hook",
			nil,
			errors.New("fail to run pre-scaling hook"),
			nil,
			&config.Config{
				PreScale: &config.Method{
					Type: "test",
				},
			},
			&fake.Execute{
				ExecuteWithValueReactor: func(method *config.Method, value string) (string, error) {
					return "", errors.New("fail to run pre-scaling hook")
				},
			},
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					replicas := int32(0)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Fail to run post-scaling hook",
			nil,
			errors.New("fail to run post-scaling hook"),
			nil,
			&config.Config{
				PostScale: &config.Method{
					Type: "test",
				},
			},
			&fake.Execute{
				ExecuteWithValueReactor: func(method *config.Method, value string) (string, error) {
					return "", errors.New("fail to run post-scaling hook")
				},
			},
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					replicas := int32(3)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Success, deployment, autoscaling disabled",
			&evaluate.Evaluation{
				TargetReplicas: 0,
			},
			nil,
			nil,
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					replicas := int32(0)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Success, deployment, autoscaling disabled, run pre-scaling hook",
			&evaluate.Evaluation{
				TargetReplicas: 0,
			},
			nil,
			nil,
			&config.Config{
				PreScale: &config.Method{
					Type: "test",
				},
			},
			&fake.Execute{
				ExecuteWithValueReactor: func(method *config.Method, value string) (string, error) {
					return "success", nil
				},
			},
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					replicas := int32(0)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Success, deployment, no change in scale",
			&evaluate.Evaluation{
				TargetReplicas: 3,
			},
			nil,
			nil,
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					replicas := int32(3)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Success, deployment, evaluation above max replicas, scale to max replicas",
			&evaluate.Evaluation{
				TargetReplicas: 5,
			},
			nil,
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
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(10),
				},
				Resource: func() *appsv1.Deployment {
					replicas := int32(2)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 5,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Success, deployment, evaluation below min replicas, scale to min replicas",
			&evaluate.Evaluation{
				TargetReplicas: 2,
			},
			nil,
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
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(1),
				},
				Resource: func() *appsv1.Deployment {
					replicas := int32(5)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 2,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Success, deployment, evaluation within min-max bounds, scale to evaluation",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
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
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *appsv1.Deployment {
					replicas := int32(5)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Success, deployment, evaluation within min-max bounds, scale to evaluation, run post-scale hook",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
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
				PostScale: &config.Method{
					Type: "test",
				},
			},
			&fake.Execute{
				ExecuteWithValueReactor: func(method *config.Method, value string) (string, error) {
					return "success", nil
				},
			},
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *appsv1.Deployment {
					replicas := int32(5)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Success, deployment, evaluation within min-max bounds, scale to evaluation, run pre and post-scale hooks",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
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
				PostScale: &config.Method{
					Type: "test",
				},
				PreScale: &config.Method{
					Type: "test",
				},
			},
			&fake.Execute{
				ExecuteWithValueReactor: func(method *config.Method, value string) (string, error) {
					return "success", nil
				},
			},
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *appsv1.Deployment {
					replicas := int32(5)
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Success, replicaset, evaluation within min-max bounds, scale to evaluation",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
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
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *appsv1.ReplicaSet {
					replicas := int32(3)
					return &appsv1.ReplicaSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.ReplicaSetSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "replicaset",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Success, replicationcontroller, evaluation within min-max bounds, scale to evaluation",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
			&scaleFake.FakeScaleClient{
				Fake: k8stesting.Fake{
					ReactionChain: []k8stesting.Reactor{
						&k8stesting.SimpleReactor{
							Resource: "replicationcontroller",
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
							Resource: "replicationcontroller",
							Verb:     "update",
							Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
								return true, &autoscaling.Scale{}, nil
							},
						},
					},
				},
			},
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *corev1.ReplicationController {
					replicas := int32(8)
					return &corev1.ReplicationController{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: corev1.ReplicationControllerSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "replicationcontroller",
					Name:       "test",
					APIVersion: "v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
		{
			"Success, statefulset, evaluation within min-max bounds, scale to evaluation",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
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
			&config.Config{},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *appsv1.StatefulSet {
					replicas := int32(1)
					return &appsv1.StatefulSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
						Spec: appsv1.StatefulSetSpec{
							Replicas: &replicas,
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "statefulset",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   autoscaler.RunType,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			scaler := &scale.Scale{
				Scaler:  test.scaler,
				Config:  test.config,
				Execute: test.executer,
			}
			result, err := scaler.Scale(test.spec)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
