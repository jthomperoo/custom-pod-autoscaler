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
	"github.com/jthomperoo/custom-pod-autoscaler/autoscaler"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/scale"
	appsv1 "k8s.io/api/apps/v1"
)

func TestScaler_Scale(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})
	var tests = []struct {
		description  string
		expected     *evaluate.Evaluation
		expectedErr  error
		spec         scale.Spec
		scaleReactor func(scale.Spec) (*evaluate.Evaluation, error)
	}{
		{
			"Return error",
			nil,
			errors.New("scale error"),
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: 4,
				},
				Resource:       &appsv1.Deployment{},
				MinReplicas:    1,
				MaxReplicas:    10,
				Namespace:      "error",
				ScaleTargetRef: nil,
				RunType:        autoscaler.RunType,
			},
			func(scale.Spec) (*evaluate.Evaluation, error) {
				return nil, errors.New("scale error")
			},
		},
		{
			"Return success",
			&evaluate.Evaluation{
				TargetReplicas: 3,
			},
			nil,
			scale.Spec{
				Evaluation: evaluate.Evaluation{
					TargetReplicas: 1,
				},
				Resource:       &appsv1.Deployment{},
				MinReplicas:    1,
				MaxReplicas:    10,
				Namespace:      "success",
				ScaleTargetRef: nil,
				RunType:        autoscaler.RunType,
			},
			func(scale.Spec) (*evaluate.Evaluation, error) {
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
