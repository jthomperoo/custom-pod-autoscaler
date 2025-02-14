/*
Copyright 2025 The Custom Pod Autoscaler Authors.

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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/resourceclient"
	meta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

func newFakeRestMapper(group string, version string, singular string, plural string) meta.RESTMapper {
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})

	groupVersion := fmt.Sprintf("%s/%s", group, version)

	mapper.AddSpecific(schema.FromAPIVersionAndKind(groupVersion, singular), schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: plural,
	}, schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: singular,
	}, meta.RESTScopeNamespace)

	return mapper
}

func TestClient_Get(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description    string
		expected       *unstructured.Unstructured
		expectedErr    error
		resourceClient resourceclient.UnstructuredClient
		apiVersion     string
		kind           string
		name           string
		namespace      string
	}{
		{
			description: "Fail to get group version resource, unknown resource",
			expected:    nil,
			expectedErr: errors.New(`no matches for kind "unknown" in version "test/v1"`),
			resourceClient: resourceclient.UnstructuredClient{
				RESTMapper: newFakeRestMapper("test", "v1", "test", "tests"),
			},
			apiVersion: "test/v1",
			kind:       "unknown",
			name:       "testname",
			namespace:  "testnamespace",
		},
		{
			description: "Fail to get resource",
			expected:    nil,
			expectedErr: errors.New(`fail to get resource`),
			resourceClient: resourceclient.UnstructuredClient{
				Dynamic: func() *fake.FakeDynamicClient {
					client := fake.NewSimpleDynamicClient(k8sruntime.NewScheme())
					client.PrependReactor("get", "tests", func(action k8stesting.Action) (bool, k8sruntime.Object, error) {
						return true, nil, errors.New("fail to get resource")
					})
					return client
				}(),
				RESTMapper: newFakeRestMapper("test", "v1", "test", "tests"),
			},
			apiVersion: "test/v1",
			kind:       "test",
			name:       "testname",
			namespace:  "testnamespace",
		},
		{
			description: "Success, Deployment",
			expected: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "testname",
						"namespace": "testnamespace",
					},
					"apiVersion": "apps/v1",
					"kind":       "deployment",
				},
			},
			expectedErr: nil,
			resourceClient: resourceclient.UnstructuredClient{
				Dynamic: fake.NewSimpleDynamicClient(k8sruntime.NewScheme(),
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "deployment",
							"metadata": map[string]interface{}{
								"namespace": "testnamespace",
								"name":      "testname",
							},
						},
					},
				),
				RESTMapper: newFakeRestMapper("apps", "v1", "deployment", "deployments"),
			},
			apiVersion: "apps/v1",
			kind:       "deployment",
			name:       "testname",
			namespace:  "testnamespace",
		},
		{
			description: "Success, Argo Rollout",
			expected: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "testname",
						"namespace": "testnamespace",
					},
					"apiVersion": "argoproj.io/v1alpha1",
					"kind":       "rollout",
				},
			},
			expectedErr: nil,
			resourceClient: resourceclient.UnstructuredClient{
				Dynamic: fake.NewSimpleDynamicClient(k8sruntime.NewScheme(),
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "argoproj.io/v1alpha1",
							"kind":       "rollout",
							"metadata": map[string]interface{}{
								"namespace": "testnamespace",
								"name":      "testname",
							},
						},
					},
				),
				RESTMapper: newFakeRestMapper("argoproj.io", "v1alpha1", "rollout", "rollouts"),
			},
			apiVersion: "argoproj.io/v1alpha1",
			kind:       "rollout",
			name:       "testname",
			namespace:  "testnamespace",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.resourceClient.Get(test.apiVersion, test.kind, test.name, test.namespace)
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
