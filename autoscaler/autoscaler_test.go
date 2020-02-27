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

package autoscaler_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/autoscaler"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/resourceclient"
	"github.com/jthomperoo/custom-pod-autoscaler/scale"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeGetMetric struct {
	getMetrics func(spec metric.Spec) ([]*metric.Metric, error)
}

func (m *fakeGetMetric) GetMetrics(spec metric.Spec) ([]*metric.Metric, error) {
	return m.getMetrics(spec)
}

type fakeGetEvaluation struct {
	getEvaluation func(spec evaluate.Spec) (*evaluate.Evaluation, error)
}

func (e *fakeGetEvaluation) GetEvaluation(spec evaluate.Spec) (*evaluate.Evaluation, error) {
	return e.getEvaluation(spec)
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
		scaler          scale.Scaler
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
				getMetric.getMetrics = func(spec metric.Spec) ([]*metric.Metric, error) {
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
				getMetric.getMetrics = func(spec metric.Spec) ([]*metric.Metric, error) {
					return []*metric.Metric{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(spec evaluate.Spec) (*evaluate.Evaluation, error) {
					return nil, errors.New("fail to evaluate")
				}
				return &getEvaluation
			}(),
		},
		{
			"Scale fail",
			errors.New("fail to scale"),
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
			&fake.Scaler{
				ScaleReactor: func(spec scale.Spec) (*evaluate.Evaluation, error) {
					return nil, errors.New("fail to scale")
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
				getMetric.getMetrics = func(spec metric.Spec) ([]*metric.Metric, error) {
					return []*metric.Metric{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(spec evaluate.Spec) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{}, nil
				}
				return &getEvaluation
			}(),
		},
		{
			"Successful autoscale",
			nil,
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					replicas := int32(1)
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
			&fake.Scaler{
				ScaleReactor: func(spec scale.Spec) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: 2,
					}, nil
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
				getMetric.getMetrics = func(spec metric.Spec) ([]*metric.Metric, error) {
					return []*metric.Metric{}, nil
				}
				return &getMetric
			}(),
			func() *fakeGetEvaluation {
				getEvaluation := fakeGetEvaluation{}
				getEvaluation.getEvaluation = func(spec evaluate.Spec) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{}, nil
				}
				return &getEvaluation
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			scaler := &autoscaler.Scaler{
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
