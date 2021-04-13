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

package pods

import (
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	metricsclient "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

// Metric (Pods) is a metric describing each pod in the current scale target (for example,
// transactions-processed-per-second).  The values will be averaged together before being compared to the target value.
type Metric struct {
	PodMetricsInfo metricsclient.PodMetricsInfo `json:"pod_metrics_info"`
	ReadyPodCount  int64                        `json:"ready_pod_count"`
	IgnoredPods    sets.String                  `json:"ignored_pods"`
	MissingPods    sets.String                  `json:"missing_pods"`
	TotalPods      int                          `json:"total_pods"`
	Timestamp      time.Time                    `json:"timestamp,omitempty"`
}
