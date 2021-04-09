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
	autoscaling "k8s.io/api/autoscaling/v1"
)

const (
	// PerPodRunMode runs metric gathering per Pod, individually running the script for each Pod being managed
	// with the Pod information piped into the metric gathering script
	PerPodRunMode = "per-pod"
	// PerResourceRunMode runs metric gathering per Resource, running the script only once for the resource
	// being managed, with the resource information piped into the metric gathering script
	PerResourceRunMode = "per-resource"
)

const (
	// APIRunType marks the metric gathering/evaluation as running during an API request, which will use the results to
	// scale
	APIRunType = "api"
	// APIDryRunRunType marks the metric gathering/evaluation as running during an API request, which will only view
	// the results and not use it for scaling
	APIDryRunRunType = "api_dry_run"
	// ScalerRunType marks the metric gathering/evaluation as running during a scale
	ScalerRunType = "scaler"
)

// Config is the configuration options for the CPA
type Config struct {
	ScaleTargetRef           *autoscaling.CrossVersionObjectReference `json:"scaleTargetRef"`
	PreMetric                *Method                                  `json:"preMetric"`
	PostMetric               *Method                                  `json:"postMetric"`
	PreEvaluate              *Method                                  `json:"preEvaluate"`
	PostEvaluate             *Method                                  `json:"postEvaluate"`
	PreScale                 *Method                                  `json:"preScale"`
	PostScale                *Method                                  `json:"postScale"`
	Evaluate                 *Method                                  `json:"evaluate"`
	Metric                   *Method                                  `json:"metric"`
	Interval                 int                                      `json:"interval"`
	Namespace                string                                   `json:"namespace"`
	MinReplicas              int32                                    `json:"minReplicas"`
	MaxReplicas              int32                                    `json:"maxReplicas"`
	RunMode                  string                                   `json:"runMode"`
	StartTime                int64                                    `json:"startTime"`
	LogVerbosity             int32                                    `json:"logVerbosity"`
	DownscaleStabilization   int                                      `json:"downscaleStabilization"`
	APIConfig                *APIConfig                               `json:"apiConfig"`
	KubernetesMetricSpecs    []K8sMetricSpec                          `json:"kubernetesMetricSpecs"`
	RequireKubernetesMetrics bool                                     `json:"requireKubernetesMetrics"`
	InitialReadinessDelay    int64                                    `json:"initialReadinessDelay"`
	CPUInitializationPeriod  int64                                    `json:"cpuInitializationPeriod"`
}
