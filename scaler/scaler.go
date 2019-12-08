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
	"log"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

// RunType scaler marks the metric gathering/evaluation as running during a scale
const RunType = "scaler"

type getMetricer interface {
	GetMetrics(deployment *appsv1.Deployment) (*metric.ResourceMetrics, error)
}

type getEvaluationer interface {
	GetEvaluation(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error)
}

// Scaler handles scaling up/down the resource being managed; triggering metric gathering and
// feeding an evaluator these metrics, before taking the results and using them to interact with Kubernetes
// to scale up/down
type Scaler struct {
	Clientset         kubernetes.Interface
	DeploymentsClient v1.DeploymentInterface
	Config            *config.Config
	GetMetricer       getMetricer
	GetEvaluationer   getEvaluationer
}

// Scale gets the managed resource, gathers metrics, evaluates these metrics and finally if a change is required
// then it scales the resource
func (s *Scaler) Scale() error {
	// Get deployment being managed
	deployment, err := s.Clientset.AppsV1().Deployments(s.Config.Namespace).Get(s.Config.ScaleTargetRef.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	currentReplicas := int32(1)
	if deployment.Spec.Replicas != nil {
		currentReplicas = *deployment.Spec.Replicas
	}

	if currentReplicas == 0 {
		log.Printf("No scaling, autoscaling disabled on resource %s", deployment.Name)
		return nil
	}

	// Gather metrics
	metrics, err := s.GetMetricer.GetMetrics(deployment)
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
	// If the deployment needs scaled up/down
	if targetReplicas != currentReplicas {
		log.Printf("Rescaling to %d replicas", targetReplicas)
		// Scale deployment
		deployment.Spec.Replicas = &targetReplicas
		_, err = s.DeploymentsClient.Update(deployment)
		if err != nil {
			return err
		}
		return nil
	}
	log.Printf("No change in target replicas, maintaining %d replicas", currentReplicas)
	return nil
}
