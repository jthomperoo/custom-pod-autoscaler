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

package resourceclient

import (
	"context"
	"fmt"

	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Client provides methods for retrieving arbitrary Kubernetes resources, returned as generalised metav1.Object, which can be converted
// to concrete types, and allows for retrieving common and shared data (namespaces, names etc.)
type Client interface {
	Get(apiVersion string, kind string, name string, namespace string) (metav1.Object, error)
}

// UnstructuredClient is an implementation of the arbitrary resource client that uses a dynamic Kubernetes interface, retrieving
// unstructured k8s objects and converting them to metav1.Object
type UnstructuredClient struct {
	Scheme                *runtime.Scheme
	Dynamic               dynamic.Interface
	UnstructuredConverter runtime.UnstructuredConverter
}

// Get takes descriptors of a Kubernetes object (api version, kind, name, namespace) and fetches the matching object, returning it
// as a metav1.Object
func (u *UnstructuredClient) Get(apiVersion string, kind string, name string, namespace string) (metav1.Object, error) {
	// TODO: update this to be less hacky
	// Convert to plural and lowercase
	kindPlural := fmt.Sprintf("%ss", strings.ToLower(kind))

	// Parse group version
	resourceGV, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, err
	}

	// Build GVK
	resourceGVK := schema.FromAPIVersionAndKind(apiVersion, kind)

	// Build GVR
	resourceGVR := schema.GroupVersionResource{
		Group:    resourceGV.Group,
		Version:  resourceGV.Version,
		Resource: kindPlural,
	}

	// Get resource
	resource, err := u.Dynamic.Resource(resourceGVR).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	resourceMeta, err := u.Scheme.New(resourceGVK)
	if err != nil {
		return nil, err
	}

	err = u.UnstructuredConverter.FromUnstructured(resource.Object, resourceMeta)
	if err != nil {
		return nil, err
	}

	return resourceMeta.(metav1.Object), nil
}
