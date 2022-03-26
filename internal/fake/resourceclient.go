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

package fake

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// ResourceClient (fake) allows inserting logic into a resource client for testing
type ResourceClient struct {
	GetReactor func(apiVersion string, kind string, name string, namespace string) (*unstructured.Unstructured, error)
}

// Get calls the fake ResourceClient reactor method provided
func (u *ResourceClient) Get(apiVersion string, kind string, name string, namespace string) (*unstructured.Unstructured, error) {
	return u.GetReactor(apiVersion, kind, name, namespace)
}
