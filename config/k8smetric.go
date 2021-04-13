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

package config

import (
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
)

// K8sMetricSpec defines which metrics to query from the metrics server
type K8sMetricSpec struct {
	Type     autoscaling.MetricSourceType `json:"type"`
	Object   *K8sObjectMetricSource       `json:"object,omitempty"`
	Pods     *K8sPodsMetricSource         `json:"pods,omitempty"`
	Resource *K8sResourceMetricSource     `json:"resource,omitempty"`
	External *K8sExternalMetricSource     `json:"external,omitempty"`
}

// K8sMetricTarget defines the type of metric gathering, either target value, average value, or average utilization of a
// specific metric
type K8sMetricTarget struct {
	Type autoscaling.MetricTargetType `json:"type"`
}

// K8sObjectMetricSource defines gathering metrics for a kubernetes object (for example, hits-per-second on an Ingress
// object).
type K8sObjectMetricSource struct {
	DescribedObject autoscaling.CrossVersionObjectReference `json:"describedObject"`
	Metric          autoscaling.MetricIdentifier            `json:"metric"`
	Target          K8sMetricTarget                         `json:"target"`
}

// K8sPodsMetricSource defines gathering metrics describing each pod in the current scale target (for example,
// transactions-processed-per-second).
type K8sPodsMetricSource struct {
	Metric autoscaling.MetricIdentifier `json:"metric"`
	Target K8sMetricTarget              `json:"target"`
}

// K8sResourceMetricSource defines gathering metrics for a resource metric known to Kubernetes, as specified in requests
// and limits, describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to
// Kubernetes.
type K8sResourceMetricSource struct {
	Name   v1.ResourceName `json:"name" protobuf:"bytes,1,name=name"`
	Target K8sMetricTarget `json:"target"`
}

// K8sExternalMetricSource defines gathering metrics for a metric not associated with any Kubernetes object (for example
// length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).
type K8sExternalMetricSource struct {
	Metric autoscaling.MetricIdentifier `json:"metric"`
	Target K8sMetricTarget              `json:"target"`
}
