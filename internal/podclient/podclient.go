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

// Package podclient provides an on-demand client for retrieving pods, without
// using caching, as the HorizontalPodAutoscaler does.
package podclient

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// OnDemandPodNamespaceLister is used to list Pods/get a specific pod in a namespace
type OnDemandPodNamespaceLister struct {
	Namespace string
	Clientset kubernetes.Interface
}

// List lists pods that match the selector in the namespace
func (p *OnDemandPodNamespaceLister) List(selector labels.Selector) ([]*corev1.Pod, error) {
	pods, err := p.Clientset.CoreV1().Pods(p.Namespace).List(context.Background(), v1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	var podPointers []*corev1.Pod
	for i := 0; i < len(pods.Items); i++ {
		podPointers = append(podPointers, &pods.Items[i])
	}
	return podPointers, nil
}

// Get gets a single pod with the name provided in the namespace
func (p *OnDemandPodNamespaceLister) Get(name string) (*corev1.Pod, error) {
	pod, err := p.Clientset.CoreV1().Pods(p.Namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

// OnDemandPodLister is used to list Pods across a cluster or retrieve a Namespaced Pod Lister
type OnDemandPodLister struct {
	Clientset kubernetes.Interface
}

// List lists pods that match the selector across the cluster
func (p *OnDemandPodLister) List(selector labels.Selector) ([]*corev1.Pod, error) {
	pods, err := p.Clientset.CoreV1().Pods("").List(context.Background(), v1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	var podPointers []*corev1.Pod
	for i := 0; i < len(pods.Items); i++ {
		podPointers = append(podPointers, &pods.Items[i])
	}
	return podPointers, nil
}

// Pods returns a namespaced pod lister in the namespace provided
func (p *OnDemandPodLister) Pods(namespace string) corelisters.PodNamespaceLister {
	return &OnDemandPodNamespaceLister{
		Namespace: namespace,
		Clientset: p.Clientset,
	}
}
