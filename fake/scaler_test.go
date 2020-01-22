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

package fake_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestScaler_Scale(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})
	var tests = []struct {
		description    string
		expected       *evaluate.Evaluation
		expectedErr    error
		evaluation     evaluate.Evaluation
		resource       metav1.Object
		minReplicas    int32
		maxReplicas    int32
		scaleTargetRef *autoscaling.CrossVersionObjectReference
		namespace      string
		scaleReactor   func(evaluation evaluate.Evaluation, resource metav1.Object, minReplicas int32, maxReplicas int32, scaleTargetRef *autoscaling.CrossVersionObjectReference, namespace string) (*evaluate.Evaluation, error)
	}{
		{
			"Return error",
			nil,
			errors.New("scale error"),
			evaluate.Evaluation{
				TargetReplicas: 4,
			},
			&appsv1.Deployment{},
			1,
			10,
			nil,
			"error",
			func(evaluation evaluate.Evaluation, resource metav1.Object, minReplicas, maxReplicas int32, scaleTargetRef *autoscaling.CrossVersionObjectReference, namespace string) (*evaluate.Evaluation, error) {
				return nil, errors.New("scale error")
			},
		},
		{
			"Return success",
			&evaluate.Evaluation{
				TargetReplicas: 3,
			},
			nil,
			evaluate.Evaluation{
				TargetReplicas: 1,
			},
			&appsv1.Deployment{},
			1,
			10,
			nil,
			"success",
			func(evaluation evaluate.Evaluation, resource metav1.Object, minReplicas, maxReplicas int32, scaleTargetRef *autoscaling.CrossVersionObjectReference, namespace string) (*evaluate.Evaluation, error) {
				return &evaluate.Evaluation{
					TargetReplicas: 3,
				}, nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			scaler := &fake.Scaler{
				ScaleReactor: test.scaleReactor,
			}
			result, err := scaler.Scale(test.evaluation, test.resource, test.minReplicas, test.maxReplicas, test.scaleTargetRef, test.namespace)
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
