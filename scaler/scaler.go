/*
Copyright 2019 The Custom Pod Autoscaler Authors.

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

// Package scaler provides methods for scaling a resource - by triggering metric
// gathering, feeding these metrics to an evaluation and using this evaluation
// to scale the resource. Handles interactions with Kubernetes API for scaling.
package scaler

import (
	"fmt"
	"log"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/resourceclient"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale"
)

// RunType scaler marks the metric gathering/evaluation as running during a scale
const RunType = "scaler"

// Scaler handles scaling up/down the resource being managed; triggering metric gathering and
// feeding an evaluator these metrics, before taking the results and using them to interact with Kubernetes
// to scale up/down
type Scaler struct {
	Scaler          scale.ScalesGetter
	Client          resourceclient.Client
	Config          *config.Config
	GetMetricer     metric.GetMetricer
	GetEvaluationer evaluate.GetEvaluationer
}

// Scale gets the managed resource, gathers metrics, evaluates these metrics and finally if a change is required
// then it scales the resource
func (s *Scaler) Scale() error {
	// Get resource being managed
	resource, err := s.Client.Get(s.Config.ScaleTargetRef.APIVersion, s.Config.ScaleTargetRef.Kind, s.Config.ScaleTargetRef.Name, s.Config.Namespace)
	if err != nil {
		return err
	}

	// Get replica count
	currentReplicas := int32(1)
	resourceReplicas, err := s.getReplicaCount(resource)
	if err != nil {
		return err
	}
	if resourceReplicas != nil {
		currentReplicas = *resourceReplicas
	}

	if currentReplicas == 0 {
		log.Printf("No scaling, autoscaling disabled on resource %s", resource.GetName())
		return nil
	}

	// Gather metrics
	metrics, err := s.GetMetricer.GetMetrics(resource)
	if err != nil {
		return err
	}

	// Mark the runtype as scaler
	metrics.RunType = RunType

	// Evaluate based on metrics
	evaluation, err := s.GetEvaluationer.GetEvaluation(metrics)
	if err != nil {
		return err
	}

	targetReplicas := currentReplicas
	if evaluation.TargetReplicas < s.Config.MinReplicas {
		log.Printf("Scale target less than min at %d replicas, setting target to min %d replicas", evaluation.TargetReplicas, s.Config.MinReplicas)
		targetReplicas = s.Config.MinReplicas
	} else if evaluation.TargetReplicas > s.Config.MaxReplicas {
		log.Printf("Scale target greater than max at %d replicas, setting target to max %d replicas", evaluation.TargetReplicas, s.Config.MaxReplicas)
		targetReplicas = s.Config.MaxReplicas
	} else {
		log.Printf("Scale target set to %d replicas", evaluation.TargetReplicas)
		targetReplicas = evaluation.TargetReplicas
	}

	// Check if evaluation requires an action
	// If the resource needs scaled up/down
	if targetReplicas != currentReplicas {
		log.Printf("Rescaling to %d replicas", targetReplicas)
		// Parse group version
		resourceGV, err := schema.ParseGroupVersion(s.Config.ScaleTargetRef.APIVersion)
		if err != nil {
			return err
		}

		targetGR := schema.GroupResource{
			Group:    resourceGV.Group,
			Resource: s.Config.ScaleTargetRef.Kind,
		}

		// Get scale for resource
		scale, err := s.Scaler.Scales(s.Config.Namespace).Get(targetGR, s.Config.ScaleTargetRef.Name)
		if err != nil {
			return err
		}

		// Scale resource
		scale.Spec.Replicas = targetReplicas
		_, err = s.Scaler.Scales(s.Config.Namespace).Update(targetGR, scale)
		if err != nil {
			return err
		}
		return nil
	}
	log.Printf("No change in target replicas, maintaining %d replicas", currentReplicas)
	return nil
}

func (s *Scaler) getReplicaCount(resource metav1.Object) (*int32, error) {
	switch v := resource.(type) {
	case *appsv1.Deployment:
		return v.Spec.Replicas, nil
	case *appsv1.ReplicaSet:
		return v.Spec.Replicas, nil
	case *appsv1.StatefulSet:
		return v.Spec.Replicas, nil
	case *corev1.ReplicationController:
		return v.Spec.Replicas, nil
	default:
		return nil, fmt.Errorf("Unsupported resource of type %T", v)
	}
}
