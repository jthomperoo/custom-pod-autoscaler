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

package metric

import (
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceMetric is the result of the custom metric calculation, containing information on the
// relevant resource and the metric value
type ResourceMetric struct {
	Resource string `json:"resource,omitempty"`
	Value    string `json:"value,omitempty"`
}

// Info defines information fed into a gatherer to retrieve metrics, contains an optional
// field 'Metrics' for storing the result
type Info struct {
	Resource          metav1.Object       `json:"resource"`
	Metrics           []*ResourceMetric   `json:"metrics,omitempty"`
	RunType           string              `json:"runType"`
	KubernetesMetrics []*k8smetric.Metric `json:"kubernetesMetrics,omitempty"`
}
