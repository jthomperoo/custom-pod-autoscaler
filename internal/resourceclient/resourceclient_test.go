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

package resourceclient_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/resourceclient"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestClient_Get(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description           string
		expected              *unstructured.Unstructured
		expectedErr           error
		dynamic               dynamic.Interface
		unstructuredConverter k8sruntime.UnstructuredConverter
		apiVersion            string
		kind                  string
		name                  string
		namespace             string
	}{
		{
			"Invalid group version, fail to parse",
			nil,
			errors.New(`unexpected GroupVersion string: /invalid/`),
			nil,
			nil,
			"/invalid/",
			"",
			"",
			"",
		},
		{
			"Fail to get resource",
			nil,
			errors.New(`fail to get resource`),
			func() *fake.FakeDynamicClient {
				client := fake.NewSimpleDynamicClient(k8sruntime.NewScheme())
				client.PrependReactor("get", "tests", func(action k8stesting.Action) (bool, k8sruntime.Object, error) {
					return true, nil, errors.New("fail to get resource")
				})
				return client
			}(),
			nil,
			"test/v1",
			"test",
			"testname",
			"testnamespace",
		},
		{
			"Success, Deployment",
			&unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "testname",
						"namespace": "testnamespace",
					},
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
				},
			},
			nil,
			fake.NewSimpleDynamicClient(k8sruntime.NewScheme(),
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "apps/v1",
						"kind":       "Deployment",
						"metadata": map[string]interface{}{
							"namespace": "testnamespace",
							"name":      "testname",
						},
					},
				},
			),
			k8sruntime.DefaultUnstructuredConverter,
			"apps/v1",
			"Deployment",
			"testname",
			"testnamespace",
		},
		{
			"Success, Argo Rollout",
			&unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "testname",
						"namespace": "testnamespace",
					},
					"apiVersion": "argoproj.io/v1alpha1",
					"kind":       "Rollout",
				},
			},
			nil,
			fake.NewSimpleDynamicClient(k8sruntime.NewScheme(),
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "argoproj.io/v1alpha1",
						"kind":       "Rollout",
						"metadata": map[string]interface{}{
							"namespace": "testnamespace",
							"name":      "testname",
						},
					},
				},
			),
			k8sruntime.DefaultUnstructuredConverter,
			"argoproj.io/v1alpha1",
			"Rollout",
			"testname",
			"testnamespace",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			scaler := &resourceclient.UnstructuredClient{
				Dynamic: test.dynamic,
			}
			result, err := scaler.Get(test.apiVersion, test.kind, test.name, test.namespace)
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
