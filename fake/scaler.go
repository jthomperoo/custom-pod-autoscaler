/*
Copyright 2020 The Custom Pod Autoscaler Authors.

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
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Scaler (fake) allows inserting logic into a scaler for testing
type Scaler struct {
	ScaleReactor func(evaluation evaluate.Evaluation, resource metav1.Object, minReplicas int32, maxReplicas int32, scaleTargetRef *autoscaling.CrossVersionObjectReference, namespace string) (*evaluate.Evaluation, error)
}

// Scale calls the fake Scaler reactor method provided
func (s *Scaler) Scale(evaluation evaluate.Evaluation, resource metav1.Object, minReplicas int32, maxReplicas int32, scaleTargetRef *autoscaling.CrossVersionObjectReference, namespace string) (*evaluate.Evaluation, error) {
	return s.ScaleReactor(evaluation, resource, minReplicas, maxReplicas, scaleTargetRef, namespace)
}
