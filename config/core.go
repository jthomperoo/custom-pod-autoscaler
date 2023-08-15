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
	autoscaling "k8s.io/api/autoscaling/v2"
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

const (
	// DefaultInterval is the default interval value
	DefaultInterval = 15000
	// DefaultNamespace is the default namespace value
	DefaultNamespace = "default"
	// DefaultMinReplicas is the default minimum replica count
	DefaultMinReplicas = 1
	// DefaultMaxReplicas is the default maximum replica count
	DefaultMaxReplicas = 10
	// DefaultStartTime is the default start time
	DefaultStartTime = 1
	// DefaultRunMode is the default run mode
	DefaultRunMode = PerPodRunMode
	// DefaultLogVerbosity is the default log verbosity
	DefaultLogVerbosity = 0
	// DefaultDownscaleStabilization is the default downscale stabilization value
	DefaultDownscaleStabilization = 0
	// DefaultCPUInitializationPeriod is the default CPU initialization value
	DefaultCPUInitializationPeriod = 300
	// DefaultInitialReadinessDelay is the default initial readiness delay value
	DefaultInitialReadinessDelay = 30
)

const (
	// DefaultAPIEnabled is the default value for the API being enabled
	DefaultAPIEnabled = true
	// DefaultUseHTTPS is the default value for the API using HTTPS
	DefaultUseHTTPS = false
	// DefaultHost is the default address for the API
	DefaultHost = "0.0.0.0"
	// DefaultPort is the default port for the API
	DefaultPort = 5000
	// DefaultCertFile is the default cert file for the API
	DefaultCertFile = ""
	// DefaultKeyFile is the default private key file for the API
	DefaultKeyFile = ""
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

// NewConfig returns a config set up with all default values
func NewConfig() *Config {
	return &Config{
		Interval:               DefaultInterval,
		Namespace:              DefaultNamespace,
		MinReplicas:            DefaultMinReplicas,
		MaxReplicas:            DefaultMaxReplicas,
		StartTime:              DefaultStartTime,
		RunMode:                DefaultRunMode,
		DownscaleStabilization: DefaultDownscaleStabilization,
		APIConfig: &APIConfig{
			Enabled:  DefaultAPIEnabled,
			UseHTTPS: DefaultUseHTTPS,
			Port:     DefaultPort,
			Host:     DefaultHost,
			CertFile: DefaultCertFile,
			KeyFile:  DefaultKeyFile,
		},
		KubernetesMetricSpecs:    nil,
		RequireKubernetesMetrics: false,
		InitialReadinessDelay:    DefaultInitialReadinessDelay,
		CPUInitializationPeriod:  DefaultCPUInitializationPeriod,
	}
}
