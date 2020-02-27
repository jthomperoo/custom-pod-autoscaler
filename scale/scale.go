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

// Package scale abstracts interactions with the Kubernetes scale API, providing a consistent way to scale
// resources that are supported by the Custom Pod Autoscaler.
package scale

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/execute"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale"
)

// Scaler abstracts interactions with the Kubernetes scale API, allowing scaling based on an evaluation provided
type Scaler interface {
	Scale(evaluation evaluate.Evaluation, resource metav1.Object, minReplicas int32, maxReplicas int32, scaleTargetRef *autoscaling.CrossVersionObjectReference, namespace string) (*evaluate.Evaluation, error)
}

// Scale interacts with the Kubernetes API to allow scaling on evaluations
type Scale struct {
	Scaler  scale.ScalesGetter
	Config  *config.Config
	Execute execute.Executer
}

// Scale takes an evaluation and uses it to interact with the Kubernetes scaling API,
// to scale up/down, or keep the same number of replicas for a resource
func (s *Scale) Scale(evaluation evaluate.Evaluation, resource metav1.Object, minReplicas int32, maxReplicas int32, scaleTargetRef *autoscaling.CrossVersionObjectReference, namespace string) (*evaluate.Evaluation, error) {
	glog.V(3).Infof("Determining replica count for resource '%s'", resource.GetName())
	currentReplicas := int32(1)
	resourceReplicas, err := s.getReplicaCount(resource)
	if err != nil {
		return nil, err
	}
	if resourceReplicas != nil {
		currentReplicas = *resourceReplicas
	}
	glog.V(3).Infof("Replica count determined: %d", currentReplicas)

	targetReplicas := currentReplicas
	if evaluation.TargetReplicas < minReplicas {
		glog.V(1).Infof("Scale target less than min at %d replicas, setting target to min %d replicas", targetReplicas, minReplicas)
		targetReplicas = minReplicas
	} else if evaluation.TargetReplicas > maxReplicas {
		glog.V(1).Infof("Scale target greater than max at %d replicas, setting target to max %d replicas", targetReplicas, minReplicas)
		targetReplicas = maxReplicas
	} else {
		glog.V(1).Infof("Scale target set to %d replicas", targetReplicas)
		targetReplicas = evaluation.TargetReplicas
	}

	scalingHookInput := struct {
		MinReplicas     int32         `json:"minReplicas"`
		MaxReplicas     int32         `json:"maxReplicas"`
		CurrentReplicas int32         `json:"currentReplicas"`
		TargetReplicas  int32         `json:"targetReplicas"`
		Resource        metav1.Object `json:"resource"`
	}{
		MinReplicas:     minReplicas,
		MaxReplicas:     maxReplicas,
		CurrentReplicas: currentReplicas,
		TargetReplicas:  targetReplicas,
		Resource:        resource,
	}

	// Convert scaling hook input to JSON
	scalingHookJSON, err := json.Marshal(scalingHookInput)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	if s.Config.PreScale != nil {
		glog.V(3).Infoln("Attempting to run pre-scaling hook")
		hookResult, err := s.Execute.ExecuteWithValue(s.Config.PreScale, string(scalingHookJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Pre-scaling hook response: %+v", hookResult)
	}

	if currentReplicas == 0 {
		glog.V(0).Infof("No scaling, autoscaling disabled on resource %s", resource.GetName())
		evaluation.TargetReplicas = 0
		return &evaluation, nil
	}

	evaluation.TargetReplicas = targetReplicas

	// Check if evaluation requires an action
	// If the resource needs scaled up/down
	if targetReplicas != currentReplicas {
		glog.V(0).Infof("Rescaling from %d to %d replicas", currentReplicas, targetReplicas)
		glog.V(3).Infoln("Attempting to parse group version")
		// Parse group version
		resourceGV, err := schema.ParseGroupVersion(scaleTargetRef.APIVersion)
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Group version parsed: %+v", resourceGV)

		targetGR := schema.GroupResource{
			Group:    resourceGV.Group,
			Resource: scaleTargetRef.Kind,
		}

		glog.V(3).Infoln("Attempting to get scale subresource for managed resource")
		scale, err := s.Scaler.Scales(namespace).Get(targetGR, scaleTargetRef.Name)
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Scale subresource retrieved: %+v", scale)

		glog.V(3).Infoln("Attempting to apply scaling changes to resource")
		scale.Spec.Replicas = targetReplicas
		_, err = s.Scaler.Scales(namespace).Update(targetGR, scale)
		if err != nil {
			return nil, err
		}
		glog.V(3).Infoln("Application of scale successful")
	} else {
		glog.V(0).Infof("No change in target replicas, maintaining %d replicas", currentReplicas)
	}

	if s.Config.PostScale != nil {
		glog.V(3).Infoln("Attempting to run post-scaling hook")
		hookResult, err := s.Execute.ExecuteWithValue(s.Config.PostScale, string(scalingHookJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Post-scaling hook response: %+v", hookResult)
	}

	return &evaluation, nil
}

func (s *Scale) getReplicaCount(resource metav1.Object) (*int32, error) {
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
