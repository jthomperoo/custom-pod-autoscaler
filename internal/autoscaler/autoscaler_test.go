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

package autoscaler_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/autoscaler"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/scale"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type fakeGetMetric struct {
	getMetrics func(info metric.Info, scaleResource *autoscalingv1.Scale) ([]*metric.ResourceMetric, error)
}

func (m *fakeGetMetric) GetMetrics(info metric.Info, scaleResource *autoscalingv1.Scale) ([]*metric.ResourceMetric, error) {
	return m.getMetrics(info, scaleResource)
}

type fakeGetEvaluation struct {
	getEvaluation func(info evaluate.Info) (*evaluate.Evaluation, error)
}

func (e *fakeGetEvaluation) GetEvaluation(info evaluate.Info) (*evaluate.Evaluation, error) {
	return e.getEvaluation(info)
}

func TestScaler(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    error
		scaler      autoscaler.Scaler
	}{
		{
			"Get resource fail",
			errors.New(`failed to get managed resource: fail to get resource`),
			autoscaler.Scaler{
				Client: &fake.ResourceClient{
					GetReactor: func(apiVersion, kind, name, namespace string) (*unstructured.Unstructured, error) {
						return nil, errors.New("fail to get resource")
					},
				},
				Config: &config.Config{
					Namespace: "test namespace",
					ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
						Kind:       "deployment",
						Name:       "test",
						APIVersion: "apps/v1",
					},
				},
			},
		},
		{
			"Gather metric fail",
			errors.New("failed to get metrics: fail to get metric"),
			autoscaler.Scaler{
				Client: &fake.ResourceClient{
					GetReactor: func(apiVersion, kind, name, namespace string) (*unstructured.Unstructured, error) {
						return &unstructured.Unstructured{
							Object: map[string]interface{}{
								"metadata": map[string]interface{}{
									"name":      name,
									"namespace": namespace,
								},
							},
						}, nil
					},
				},
				Config: &config.Config{
					Namespace: "test namespace",
					ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
						Kind:       "deployment",
						Name:       "test",
						APIVersion: "apps/v1",
					},
				},
				GetMetricer: func() *fakeGetMetric {
					getMetric := fakeGetMetric{}
					getMetric.getMetrics = func(spec metric.Info, scaleResource *autoscalingv1.Scale) ([]*metric.ResourceMetric, error) {
						return nil, errors.New("fail to get metric")
					}
					return &getMetric
				}(),
				Scaler: &fake.Scaler{
					GetScaleSubResourceReactor: func(apiVersion, kind, namespace, name string) (*autoscalingv1.Scale, error) {
						return &autoscalingv1.Scale{
							Spec: autoscalingv1.ScaleSpec{
								Replicas: 1,
							},
						}, nil
					},
				},
			},
		},
		{
			"Evaluate fail",
			errors.New("failed get evaluation: fail to evaluate"),
			autoscaler.Scaler{
				Client: &fake.ResourceClient{
					GetReactor: func(apiVersion, kind, name, namespace string) (*unstructured.Unstructured, error) {
						return &unstructured.Unstructured{
							Object: map[string]interface{}{
								"metadata": map[string]interface{}{
									"name":      name,
									"namespace": namespace,
								},
							},
						}, nil
					},
				},
				Config: &config.Config{
					Namespace: "test namespace",
					ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
						Kind:       "deployment",
						Name:       "test",
						APIVersion: "apps/v1",
					},
				},
				GetMetricer: func() *fakeGetMetric {
					getMetric := fakeGetMetric{}
					getMetric.getMetrics = func(spec metric.Info, scaleResource *autoscalingv1.Scale) ([]*metric.ResourceMetric, error) {
						return []*metric.ResourceMetric{}, nil
					}
					return &getMetric
				}(),
				GetEvaluationer: func() *fakeGetEvaluation {
					getEvaluation := fakeGetEvaluation{}
					getEvaluation.getEvaluation = func(info evaluate.Info) (*evaluate.Evaluation, error) {
						return nil, errors.New("fail to evaluate")
					}
					return &getEvaluation
				}(),
				Scaler: &fake.Scaler{
					GetScaleSubResourceReactor: func(apiVersion, kind, namespace, name string) (*autoscalingv1.Scale, error) {
						return &autoscalingv1.Scale{
							Spec: autoscalingv1.ScaleSpec{
								Replicas: 1,
							},
						}, nil
					},
				},
			},
		},
		{
			"Scale fail",
			errors.New("failed to scale resource: fail to scale"),
			autoscaler.Scaler{
				Client: &fake.ResourceClient{
					GetReactor: func(apiVersion, kind, name, namespace string) (*unstructured.Unstructured, error) {
						return &unstructured.Unstructured{
							Object: map[string]interface{}{
								"metadata": map[string]interface{}{
									"name":      name,
									"namespace": namespace,
								},
								"spec": map[string]interface{}{
									"replicas": 2,
								},
							},
						}, nil
					},
				},
				Scaler: &fake.Scaler{
					ScaleReactor: func(info scale.Info, scaleResource *autoscalingv1.Scale) (*evaluate.Evaluation, error) {
						return nil, errors.New("fail to scale")
					},
					GetScaleSubResourceReactor: func(apiVersion, kind, namespace, name string) (*autoscalingv1.Scale, error) {
						return &autoscalingv1.Scale{
							Spec: autoscalingv1.ScaleSpec{
								Replicas: 1,
							},
						}, nil
					},
				},
				Config: &config.Config{
					Namespace: "test namespace",
					ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
						Kind:       "deployment",
						Name:       "test",
						APIVersion: "apps/v1",
					},
				},
				GetMetricer: func() *fakeGetMetric {
					getMetric := fakeGetMetric{}
					getMetric.getMetrics = func(info metric.Info, scaleResource *autoscalingv1.Scale) ([]*metric.ResourceMetric, error) {
						return []*metric.ResourceMetric{}, nil
					}
					return &getMetric
				}(),
				GetEvaluationer: func() *fakeGetEvaluation {
					getEvaluation := fakeGetEvaluation{}
					getEvaluation.getEvaluation = func(info evaluate.Info) (*evaluate.Evaluation, error) {
						return &evaluate.Evaluation{}, nil
					}
					return &getEvaluation
				}(),
			},
		},
		{
			"Successful autoscale",
			nil,
			autoscaler.Scaler{
				Client: &fake.ResourceClient{
					GetReactor: func(apiVersion, kind, name, namespace string) (*unstructured.Unstructured, error) {
						return &unstructured.Unstructured{
							Object: map[string]interface{}{
								"metadata": map[string]interface{}{
									"name":      name,
									"namespace": namespace,
								},
								"spec": map[string]interface{}{
									"replicas": 1,
								},
							},
						}, nil
					},
				},
				Scaler: &fake.Scaler{
					ScaleReactor: func(info scale.Info, scaleResource *autoscalingv1.Scale) (*evaluate.Evaluation, error) {
						return &evaluate.Evaluation{
							TargetReplicas: 2,
						}, nil
					},
					GetScaleSubResourceReactor: func(apiVersion, kind, namespace, name string) (*autoscalingv1.Scale, error) {
						return &autoscalingv1.Scale{
							Spec: autoscalingv1.ScaleSpec{
								Replicas: 1,
							},
						}, nil
					},
				},
				Config: &config.Config{
					Namespace: "test namespace",
					ScaleTargetRef: &autoscalingv2.CrossVersionObjectReference{
						Kind:       "deployment",
						Name:       "test",
						APIVersion: "apps/v1",
					},
				},
				GetMetricer: func() *fakeGetMetric {
					getMetric := fakeGetMetric{}
					getMetric.getMetrics = func(spec metric.Info, scaleResource *autoscalingv1.Scale) ([]*metric.ResourceMetric, error) {
						return []*metric.ResourceMetric{}, nil
					}
					return &getMetric
				}(),
				GetEvaluationer: func() *fakeGetEvaluation {
					getEvaluation := fakeGetEvaluation{}
					getEvaluation.getEvaluation = func(info evaluate.Info) (*evaluate.Evaluation, error) {
						return &evaluate.Evaluation{}, nil
					}
					return &getEvaluation
				}(),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := test.scaler.Scale()
			if !cmp.Equal(&err, &test.expected, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expected, err, equateErrorMessage))
			}
		})
	}
}
