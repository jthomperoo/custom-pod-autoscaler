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

package resourceclient_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/resourceclient"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

type fakeUnstructuredConverter struct {
	ToUnstructuredReactor   func(obj interface{}) (map[string]interface{}, error)
	FromUnstructuredReactor func(u map[string]interface{}, obj interface{}) error
}

func (f *fakeUnstructuredConverter) ToUnstructured(obj interface{}) (map[string]interface{}, error) {
	return f.ToUnstructuredReactor(obj)
}

func (f *fakeUnstructuredConverter) FromUnstructured(u map[string]interface{}, obj interface{}) error {
	return f.FromUnstructuredReactor(u, obj)
}

func TestClient_Get(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description           string
		expected              metav1.Object
		expectedErr           error
		dynamic               dynamic.Interface
		unstructuredConverter runtime.UnstructuredConverter
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
				client := fake.NewSimpleDynamicClient(runtime.NewScheme())
				client.PrependReactor("get", "tests", func(action k8stesting.Action) (bool, runtime.Object, error) {
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
			"Fail to create new object using schema",
			nil,
			errors.New(`no kind "test" is registered for version "test/v1" in scheme "k8s.io/client-go/kubernetes/scheme/register.go:69"`),
			fake.NewSimpleDynamicClient(runtime.NewScheme(),
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "test/v1",
						"kind":       "test",
						"metadata": map[string]interface{}{
							"namespace": "testnamespace",
							"name":      "testname",
						},
					},
				},
			),
			nil,
			"test/v1",
			"test",
			"testname",
			"testnamespace",
		},
		{
			"Fail to convert from unstructured to object",
			nil,
			errors.New(`fail to convert from unstructured`),
			fake.NewSimpleDynamicClient(runtime.NewScheme(),
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
			&fakeUnstructuredConverter{
				FromUnstructuredReactor: func(u map[string]interface{}, obj interface{}) error {
					return errors.New("fail to convert from unstructured")
				},
			},
			"apps/v1",
			"Deployment",
			"testname",
			"testnamespace",
		},
		{
			"Success",
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testnamespace",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
			},
			nil,
			fake.NewSimpleDynamicClient(runtime.NewScheme(),
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
			runtime.DefaultUnstructuredConverter,
			"apps/v1",
			"Deployment",
			"testname",
			"testnamespace",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			scaler := &resourceclient.UnstructuredClient{
				Dynamic:               test.dynamic,
				UnstructuredConverter: test.unstructuredConverter,
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
