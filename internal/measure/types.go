/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Modifications Copyright 2021 The Custom Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.

Modified to split up evaluations and metric gathering to work with the
Custom Pod Autoscaler framework.
Original source:
https://github.com/kubernetes/api/blob/master/autoscaling/v2beta2/types.go
*/

package measure

import (
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/external"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/object"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/pods"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/resource"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
)

// MetricSpec defines which metrics to query from the metrics server
type MetricSpec struct {
	Type              autoscaling.MetricSourceType   `json:"type"`
	Object            *ObjectMetricSource            `json:"object,omitempty"`
	Pods              *PodsMetricSource              `json:"pods,omitempty"`
	Resource          *ResourceMetricSource          `json:"resource,omitempty"`
	ContainerResource *ContainerResourceMetricSource `json:"containerResource,omitempty"`
	External          *ExternalMetricSource          `json:"external,omitempty"`
}

// Metric is a metric that has been retrieved from the K8s metrics server
type Metric struct {
	CurrentReplicas int32            `json:"current_replicas"`
	Spec            MetricSpec       `json:"spec"`
	Resource        *resource.Metric `json:"resource,omitempty"`
	Pods            *pods.Metric     `json:"pods,omitempty"`
	Object          *object.Metric   `json:"object,omitempty"`
	External        *external.Metric `json:"external,omitempty"`
}

// MetricTarget defines the type of metric gathering, either target value, average value, or average utilization of a
// specific metric
type MetricTarget struct {
	Type autoscaling.MetricTargetType `json:"type"`
}

// ObjectMetricSource defines gathering metrics for a kubernetes object (for example, hits-per-second on an Ingress
// object).
type ObjectMetricSource struct {
	DescribedObject autoscaling.CrossVersionObjectReference `json:"describedObject"`
	Metric          autoscaling.MetricIdentifier            `json:"metric"`
	Target          MetricTarget                            `json:"target"`
}

// PodsMetricSource defines gathering metrics describing each pod in the current scale target (for example,
// transactions-processed-per-second).
type PodsMetricSource struct {
	Metric autoscaling.MetricIdentifier `json:"metric"`
	Target MetricTarget                 `json:"target"`
}

// ResourceMetricSource defines gathering metrics for a resource metric known to Kubernetes, as specified in requests
// and limits, describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to
// Kubernetes.
type ResourceMetricSource struct {
	Name   v1.ResourceName `json:"name" protobuf:"bytes,1,name=name"`
	Target MetricTarget    `json:"target"`
}

// ContainerResourceMetricSource defines gathering metrics for a resource metric known to Kubernetes, as specified in
// requests and limits, describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in
// to Kubernetes.
type ContainerResourceMetricSource struct {
	Name      v1.ResourceName `json:"name"`
	Container string          `json:"container"`
	Target    MetricTarget    `json:"target"`
}

// ExternalMetricSource defines gathering metrics for a metric not associated with any Kubernetes object (for example
// length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).
type ExternalMetricSource struct {
	Metric autoscaling.MetricIdentifier `json:"metric"`
	Target MetricTarget                 `json:"target"`
}
