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
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResourceClient_Get(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})
	var tests = []struct {
		description string
		expected    metav1.Object
		expectedErr error
		apiVersion  string
		kind        string
		name        string
		namespace   string
		getReactor  func(apiVersion string, kind string, name string, namespace string) (metav1.Object, error)
	}{
		{
			"Return error",
			nil,
			errors.New("get error"),
			"error",
			"error",
			"error",
			"error",
			func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
				return nil, errors.New("get error")
			},
		},
		{
			"Return success",
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "success",
				},
			},
			nil,
			"success",
			"success",
			"success",
			"success",
			func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
				return &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "success",
					},
				}, nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			execute := &fake.ResourceClient{
				GetReactor: test.getReactor,
			}
			result, err := execute.Get(test.apiVersion, test.kind, test.name, test.namespace)
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
