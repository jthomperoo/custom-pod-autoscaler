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

package scaling_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/scaling"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/scale"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1" // Client-go uses the autoscaling/v1 api for its scaling client
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
		description   string
		expected      *evaluate.Evaluation
		expectedErr   error
		scaler        scaling.Scaler
		info          scale.Info
		scaleResource *autoscalingv1.Scale
	}{
		{
			"Fail to update scale for resource",
			nil,
			errors.New("failed to apply scaling changes to resource: fail to update resource"),
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
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
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(1),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Fail to parse invalid API version",
			nil,
			errors.New("unexpected GroupVersion string: invalid/invalid/invalid"),
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
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
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(1),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "invalid/invalid/invalid",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Fail to run pre-scaling hook",
			nil,
			errors.New("failed run pre-scaling hook: fail to run pre-scaling hook"),
			&scaling.Scale{
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
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 2,
				},
			},
		},
		{
			"Fail to run post-scaling hook",
			nil,
			errors.New("failed to run post-scaling hook: fail to run post-scaling hook"),
			&scaling.Scale{
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
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Success, deployment, autoscaling disabled",
			&evaluate.Evaluation{
				TargetReplicas: 0,
			},
			nil,
			&scaling.Scale{
				nil,
				&config.Config{},
				nil,
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 0,
				},
			},
		},
		{
			"Success, deployment, scale to 0",
			&evaluate.Evaluation{
				TargetReplicas: int32(0),
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "deployment",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
								},
							},
						},
					},
				},
				&config.Config{},
				nil,
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(0),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 0,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 5,
				},
			},
		},
		{
			"Success, deployment, scale from 0",
			&evaluate.Evaluation{
				TargetReplicas: int32(3),
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "deployment",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
								},
							},
						},
					},
				},
				&config.Config{},
				nil,
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 0,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 0,
				},
			},
		},
		{
			"Success, deployment, autoscaling disabled, run pre-scaling hook",
			&evaluate.Evaluation{
				TargetReplicas: 0,
			},
			nil,
			&scaling.Scale{
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
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 0,
				},
			},
		},
		{
			"Success, deployment, no change in scale",
			&evaluate.Evaluation{
				TargetReplicas: 3,
			},
			nil,
			&scaling.Scale{
				nil,
				&config.Config{},
				nil,
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Success, deployment, evaluation above max replicas, scale to max replicas",
			&evaluate.Evaluation{
				TargetReplicas: 5,
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "deployment",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
								},
							},
						},
					},
				},
				&config.Config{},
				nil,
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(10),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 5,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 2,
				},
			},
		},
		{
			"Success, deployment, evaluation below min replicas, scale to min replicas",
			&evaluate.Evaluation{
				TargetReplicas: 2,
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "deployment",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
								},
							},
						},
					},
				},
				&config.Config{},
				nil,
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(1),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 2,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 5,
				},
			},
		},
		{
			"Success, deployment, evaluation within min-max bounds, scale to evaluation",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "deployment",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
								},
							},
						},
					},
				},
				&config.Config{},
				nil,
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 5,
				},
			},
		},
		{
			"Success, deployment, evaluation within min-max bounds, scale to evaluation, run post-scale hook",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "deployment",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
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
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 5,
				},
			},
		},
		{
			"Success, deployment, evaluation within min-max bounds, scale to evaluation, run pre and post-scale hooks",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "deployment",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
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
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 5,
				},
			},
		},
		{
			"Success, replicaset, evaluation within min-max bounds, scale to evaluation",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "replicaset",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
								},
							},
						},
					},
				},
				&config.Config{},
				nil,
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *appsv1.ReplicaSet {
					return &appsv1.ReplicaSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "replicaset",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Success, argo rollout, evaluation within min-max bounds, scale to evaluation",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "rollout",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
								},
							},
						},
					},
				},
				&config.Config{},
				nil,
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *unstructured.Unstructured {
					return &unstructured.Unstructured{
						Object: map[string]interface{}{
							"name":      "test",
							"namespace": "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "rollout",
					Name:       "test",
					APIVersion: "argoproj.io/v1alpha1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Success, replicationcontroller, evaluation within min-max bounds, scale to evaluation",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "replicationcontroller",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
								},
							},
						},
					},
				},
				&config.Config{},
				nil,
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *corev1.ReplicationController {
					return &corev1.ReplicationController{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "replicationcontroller",
					Name:       "test",
					APIVersion: "v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 8,
				},
			},
		},
		{
			"Success, statefulset, evaluation within min-max bounds, scale to evaluation",
			&evaluate.Evaluation{
				TargetReplicas: int32(7),
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "statefulset",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
								},
							},
						},
					},
				},
				&config.Config{},
				nil,
				nil,
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(7),
				},
				Resource: func() *appsv1.StatefulSet {
					return &appsv1.StatefulSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "statefulset",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 1,
				},
			},
		},
		{
			"Success, deployment, 3 values within downscale stabilization window, 2 values pruned, previous max",
			&evaluate.Evaluation{
				TargetReplicas: int32(9),
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "deployment",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
								},
							},
						},
					},
				},
				&config.Config{
					DownscaleStabilization: 45,
				},
				nil,
				[]scaling.TimestampedEvaluation{
					{
						Time: time.Now().Add(-60 * time.Second),
						Evaluation: evaluate.Evaluation{
							TargetReplicas: 100,
						},
					},
					{
						Time: time.Now().Add(-50 * time.Second),
						Evaluation: evaluate.Evaluation{
							TargetReplicas: 2,
						},
					},
					{
						Time: time.Now().Add(-40 * time.Second),
						Evaluation: evaluate.Evaluation{
							TargetReplicas: 2,
						},
					},
					{
						Time: time.Now().Add(-30 * time.Second),
						Evaluation: evaluate.Evaluation{
							TargetReplicas: 9,
						},
					},
					{
						Time: time.Now().Add(-20 * time.Second),
						Evaluation: evaluate.Evaluation{
							TargetReplicas: 2,
						},
					},
				},
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(2),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 5,
				},
			},
		},
		{
			"Success, deployment, 3 values within downscale stabilization window, 2 values pruned, latest max",
			&evaluate.Evaluation{
				TargetReplicas: int32(3),
			},
			nil,
			&scaling.Scale{
				&scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "deployment",
								Verb:     "update",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{}, nil
								},
							},
						},
					},
				},
				&config.Config{
					DownscaleStabilization: 25,
				},
				nil,
				[]scaling.TimestampedEvaluation{
					{
						Time: time.Now().Add(-30 * time.Second),
						Evaluation: evaluate.Evaluation{
							TargetReplicas: 100,
						},
					},
					{
						Time: time.Now().Add(-20 * time.Second),
						Evaluation: evaluate.Evaluation{
							TargetReplicas: 2,
						},
					},
					{
						Time: time.Now().Add(-15 * time.Second),
						Evaluation: evaluate.Evaluation{
							TargetReplicas: 2,
						},
					},
					{
						Time: time.Now().Add(-10 * time.Second),
						Evaluation: evaluate.Evaluation{
							TargetReplicas: 1,
						},
					},
					{
						Time: time.Now().Add(-5 * time.Second),
						Evaluation: evaluate.Evaluation{
							TargetReplicas: 2,
						},
					},
				},
			},
			scale.Info{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: int32(3),
				},
				Resource: func() *appsv1.Deployment {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test namespace",
						},
					}
				}(),
				MinReplicas: 1,
				MaxReplicas: 10,
				ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
				Namespace: "test",
				RunType:   config.ScalerRunType,
			},
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 5,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.scaler.Scale(test.info, test.scaleResource)
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

func TestGetScaleSubResource(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	tests := []struct {
		name         string
		expected     *autoscalingv1.Scale
		expectedErr  error
		scale        *scaling.Scale
		apiVersion   string
		kind         string
		namespace    string
		resourceName string
	}{
		{
			name:        "Invalid api version",
			expected:    nil,
			expectedErr: fmt.Errorf("unexpected GroupVersion string: invalid/invalid/invalid"),
			scale:       &scaling.Scale{},
			apiVersion:  "invalid/invalid/invalid",
		},
		{
			name:        "Fail to get scale subresource",
			expected:    nil,
			expectedErr: fmt.Errorf("failed to get scale subresource for resource: fail to get scale subresource"),
			scale: &scaling.Scale{
				Scaler: &scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "test",
								Verb:     "get",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, nil, errors.New("fail to get scale subresource")
								},
							},
						},
					},
				},
			},
			apiVersion:   "test/v1",
			kind:         "test",
			namespace:    "test",
			resourceName: "test",
		},
		{
			name: "Success",
			expected: &autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
			expectedErr: nil,
			scale: &scaling.Scale{
				Scaler: &scaleFake.FakeScaleClient{
					Fake: k8stesting.Fake{
						ReactionChain: []k8stesting.Reactor{
							&k8stesting.SimpleReactor{
								Resource: "test",
								Verb:     "get",
								Reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
									return true, &autoscalingv1.Scale{
										Spec: autoscalingv1.ScaleSpec{
											Replicas: 3,
										},
									}, nil
								},
							},
						},
					},
				},
			},
			apiVersion:   "test/v1",
			kind:         "test",
			namespace:    "test",
			resourceName: "test",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.scale.GetScaleSubResource(test.apiVersion, test.kind, test.namespace, test.resourceName)
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
