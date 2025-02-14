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

package resourceclient

import (
	"context"

	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Client provides methods for retrieving arbitrary Kubernetes resources, returned as generalised metav1.Object, which
// can be converted to concrete types, and allows for retrieving common and shared data (namespaces, names etc.)
type Client interface {
	Get(apiVersion string, kind string, name string, namespace string) (*unstructured.Unstructured, error)
}

// UnstructuredClient is an implementation of the arbitrary resource client that uses a dynamic Kubernetes interface,
// retrieving unstructured k8s objects and converting them to metav1.Object
type UnstructuredClient struct {
	Dynamic    dynamic.Interface
	RESTMapper meta.RESTMapper
}

// Get takes descriptors of a Kubernetes object (api version, kind, name, namespace) and fetches the matching object,
// returning it as an unstructured Kubernetes resource
func (u *UnstructuredClient) Get(apiVersion string, kind string, name string, namespace string) (*unstructured.Unstructured, error) {
	resourceGK := schema.FromAPIVersionAndKind(apiVersion, kind)
	mapping, err := u.RESTMapper.RESTMapping(resourceGK.GroupKind(), resourceGK.Version)
	if err != nil {
		return nil, err
	}

	// Get resource
	resource, err := u.Dynamic.Resource(mapping.Resource).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return resource, nil
}
