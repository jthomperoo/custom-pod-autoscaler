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

	"github.com/golang/glog"
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
	glog.V(2).Infoln("Attempting to get managed resource")
	resource, err := s.Client.Get(s.Config.ScaleTargetRef.APIVersion, s.Config.ScaleTargetRef.Kind, s.Config.ScaleTargetRef.Name, s.Config.Namespace)
	if err != nil {
		return err
	}
	glog.V(2).Infof("Managed resource retrieved: %+v", resource)

	glog.V(2).Infoln("Determining replica count for resource")
	currentReplicas := int32(1)
	resourceReplicas, err := s.getReplicaCount(resource)
	if err != nil {
		return err
	}
	if resourceReplicas != nil {
		currentReplicas = *resourceReplicas
	}
	glog.V(2).Infof("Replica count determined: %d", currentReplicas)

	if currentReplicas == 0 {
		glog.V(0).Infof("No scaling, autoscaling disabled on resource %s", resource.GetName())
		return nil
	}

	glog.V(2).Infoln("Attempting to get resource metrics")
	metrics, err := s.GetMetricer.GetMetrics(resource)
	if err != nil {
		return err
	}
	glog.V(2).Infof("Retrieved metrics: %+v", metrics)

	// Mark the runtype as scaler
	metrics.RunType = RunType

	glog.V(2).Infoln("Attempting to evaluate metrics")
	evaluation, err := s.GetEvaluationer.GetEvaluation(metrics)
	if err != nil {
		return err
	}
	glog.V(2).Infof("Metrics evaluated: %+v", evaluation)

	targetReplicas := currentReplicas
	if evaluation.TargetReplicas < s.Config.MinReplicas {
		glog.V(1).Infof("Scale target less than min at %d replicas, setting target to min %d replicas", evaluation.TargetReplicas, s.Config.MinReplicas)
		targetReplicas = s.Config.MinReplicas
	} else if evaluation.TargetReplicas > s.Config.MaxReplicas {
		glog.V(1).Infof("Scale target greater than max at %d replicas, setting target to max %d replicas", evaluation.TargetReplicas, s.Config.MaxReplicas)
		targetReplicas = s.Config.MaxReplicas
	} else {
		glog.V(1).Infof("Scale target set to %d replicas", evaluation.TargetReplicas)
		targetReplicas = evaluation.TargetReplicas
	}

	// Check if evaluation requires an action
	// If the resource needs scaled up/down
	if targetReplicas != currentReplicas {
		glog.V(0).Infof("Rescaling from %d to %d replicas", currentReplicas, targetReplicas)
		glog.V(2).Infoln("Attempting to parse group version")
		// Parse group version
		resourceGV, err := schema.ParseGroupVersion(s.Config.ScaleTargetRef.APIVersion)
		if err != nil {
			return err
		}
		glog.V(2).Infof("Group version parsed: %+v", resourceGV)

		targetGR := schema.GroupResource{
			Group:    resourceGV.Group,
			Resource: s.Config.ScaleTargetRef.Kind,
		}

		glog.V(2).Infoln("Attempting to get scale subresource for managed resource")
		scale, err := s.Scaler.Scales(s.Config.Namespace).Get(targetGR, s.Config.ScaleTargetRef.Name)
		if err != nil {
			return err
		}
		glog.V(2).Infof("Scale subresource retrieved: %+v", scale)

		glog.V(2).Infoln("Attempting to apply scaling changes to resource")
		scale.Spec.Replicas = targetReplicas
		_, err = s.Scaler.Scales(s.Config.Namespace).Update(targetGR, scale)
		if err != nil {
			return err
		}
		glog.V(2).Infoln("Application of scale successful")
		return nil
	}
	glog.V(0).Infof("No change in target replicas, maintaining %d replicas", currentReplicas)
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
