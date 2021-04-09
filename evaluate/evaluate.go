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

package evaluate

import (
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Evaluation represents a decision on how to scale a resource
type Evaluation struct {
	TargetReplicas int32 `json:"targetReplicas"`
}

// Info defines information fed into an evaluator to produce an evaluation,
// contains optional 'Evaluation' field for storing the result
type Info struct {
	Metrics    []*metric.ResourceMetric `json:"metrics"`
	Resource   metav1.Object            `json:"resource"`
	Evaluation *Evaluation              `json:"evaluation,omitempty"`
	RunType    string                   `json:"runType"`
}
