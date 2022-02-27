/*
Copyright 2022 The Custom Pod Autoscaler Authors.

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

package podmetrics

import "time"

// Metric contains pod metric value (the metric values are expected to be the metric as a milli-value)
type Metric struct {
	Timestamp time.Time
	Window    time.Duration
	Value     int64
}

// MetricsInfo contains pod metrics as a map from pod names to MetricsInfo
type MetricsInfo map[string]Metric
