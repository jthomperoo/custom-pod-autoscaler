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
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

// K8sMetricSpec defines which metrics to query from the metrics server
type K8sMetricSpec autoscalingv2.MetricSpec

// K8sMetricTarget defines the type of metric gathering, either target value, average value, or average utilization of a
// specific metric
type K8sMetricTarget autoscalingv2.MetricTarget

// K8sObjectMetricSource defines gathering metrics for a kubernetes object (for example, hits-per-second on an Ingress
// object).
type K8sObjectMetricSource autoscalingv2.ObjectMetricSource

// K8sPodsMetricSource defines gathering metrics describing each pod in the current scale target (for example,
// transactions-processed-per-second).
type K8sPodsMetricSource autoscalingv2.PodsMetricSource

// K8sResourceMetricSource defines gathering metrics for a resource metric known to Kubernetes, as specified in requests
// and limits, describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to
// Kubernetes.
type K8sResourceMetricSource autoscalingv2.ResourceMetricSource

// K8sExternalMetricSource defines gathering metrics for a metric not associated with any Kubernetes object (for example
// length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).
type K8sExternalMetricSource autoscalingv2.ExternalMetricSource
