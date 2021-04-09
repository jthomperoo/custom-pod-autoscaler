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

package scale

import (
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Info defines information fed into a Scaler in order for it to make decisions as to how to scale
type Info struct {
	Evaluation     evaluate.Evaluation                      `json:"evaluation"`
	Resource       metav1.Object                            `json:"resource"`
	ScaleTargetRef *autoscaling.CrossVersionObjectReference `json:"scaleTargetRef"`
	Namespace      string                                   `json:"namespace"`
	MinReplicas    int32                                    `json:"minReplicas"`
	MaxReplicas    int32                                    `json:"maxReplicas"`
	TargetReplicas int32                                    `json:"targetReplicas"`
	RunType        string                                   `json:"runType"`
}
