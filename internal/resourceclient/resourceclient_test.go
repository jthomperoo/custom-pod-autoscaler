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
	"fmt"
	"runtime"
	"testing"

	argov1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/resourceclient"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
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

func defaultScheme() *k8sruntime.Scheme {
	scheme := k8sruntime.NewScheme()
	schemeBuilder := k8sruntime.NewSchemeBuilder(argov1alpha1.AddToScheme)
	schemeBuilder.Register(clientsetscheme.AddToScheme)
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	return scheme
}

// This is a little bit hacky, but the scheme errors print the line of the scheme that was used to create it, in these
// tests it's the function above, so line 51, but if this is changed this number will need manually updated
const schemeLineNumber = 52

func TestClient_Get(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
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
		scheme                *k8sruntime.Scheme
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
			defaultScheme(),
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
			"Fail to create new object using schema",
			nil,
			fmt.Errorf(`no kind "unknown" is registered for version "unknown/v1" in scheme "%s:%d"`, filename, schemeLineNumber),
			defaultScheme(),
			fake.NewSimpleDynamicClient(k8sruntime.NewScheme(),
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "unknown/v1",
						"kind":       "unknown",
						"metadata": map[string]interface{}{
							"namespace": "testnamespace",
							"name":      "testname",
						},
					},
				},
			),
			nil,
			"unknown/v1",
			"unknown",
			"testname",
			"testnamespace",
		},
		{
			"Fail to convert from unstructured to object",
			nil,
			errors.New(`fail to convert from unstructured`),
			defaultScheme(),
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
			"Success, Deployment",
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
			defaultScheme(),
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
			&argov1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testnamespace",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "Rollout",
					APIVersion: "argoproj.io/v1alpha1",
				},
			},
			nil,
			defaultScheme(),
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
				Scheme:  test.scheme,
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
