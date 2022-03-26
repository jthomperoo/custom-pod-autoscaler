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

package fake

import (
	"github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/scale"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
)

// Scaler (fake) allows inserting logic into a scaler for testing
type Scaler struct {
	ScaleReactor               func(info scale.Info, scaleResource *autoscalingv1.Scale) (*evaluate.Evaluation, error)
	GetScaleSubResourceReactor func(apiVersion string, kind string, namespace string, name string) (*autoscalingv1.Scale, error)
}

// Scale calls the fake Scaler reactor method provided
func (s *Scaler) Scale(info scale.Info, scaleResource *autoscalingv1.Scale) (*evaluate.Evaluation, error) {
	return s.ScaleReactor(info, scaleResource)
}

// GetScaleSubResource calls the fake GetScaleSubResource reactor method provided
func (s *Scaler) GetScaleSubResource(apiVersion string, kind string, namespace string, name string) (*autoscalingv1.Scale, error) {
	return s.GetScaleSubResourceReactor(apiVersion, kind, namespace, name)
}
