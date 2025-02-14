/*
Copyright 2025 The Custom Pod Autoscaler Authors.

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

// Package scaling abstracts interactions with the Kubernetes scale API, providing a consistent way to scale
// resources that are supported by the Custom Pod Autoscaler.
package scaling

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/execute"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/scale"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/restmapper"
	k8sscale "k8s.io/client-go/scale"
)

// Scaler abstracts interactions with the Kubernetes scale API, allowing scaling based on an evaluation provided
type Scaler interface {
	Scale(info scale.Info, scaleResource *autoscalingv1.Scale) (*evaluate.Evaluation, error)
	GetScaleSubResource(apiVersion string, kind string, namespace string, name string) (*autoscalingv1.Scale, error)
}

// Scale interacts with the Kubernetes API to allow scaling on evaluations
type Scale struct {
	Scaler                   k8sscale.ScalesGetter
	Config                   *config.Config
	Execute                  execute.Executer
	StabilizationEvaluations []TimestampedEvaluation
	RESTMapper               restmapper.DeferredDiscoveryRESTMapper
}

// TimestampedEvaluation is used to associate an evaluation with a timestamp, used in stabilizing evaluations
// with the downscale stabilization window
type TimestampedEvaluation struct {
	Time       time.Time
	Evaluation evaluate.Evaluation
}

// Scale takes an evaluation and uses it to interact with the Kubernetes scaling API,
// to scale up/down, or keep the same number of replicas for a resource
func (s *Scale) Scale(info scale.Info, scaleResource *autoscalingv1.Scale) (*evaluate.Evaluation, error) {
	currentReplicas := scaleResource.Spec.Replicas
	targetReplicas := currentReplicas
	if info.Evaluation.TargetReplicas < info.MinReplicas {
		glog.V(1).Infof("Scale target less than min at %d replicas, setting target to min %d replicas", targetReplicas, info.MinReplicas)
		targetReplicas = info.MinReplicas
	} else if info.Evaluation.TargetReplicas > info.MaxReplicas {
		glog.V(1).Infof("Scale target greater than max at %d replicas, setting target to max %d replicas", targetReplicas, info.MinReplicas)
		targetReplicas = info.MaxReplicas
	} else {
		glog.V(1).Infof("Scale target set to %d replicas", targetReplicas)
		targetReplicas = info.Evaluation.TargetReplicas
	}

	if currentReplicas == 0 && info.MinReplicas != 0 {
		glog.V(0).Infof("No scaling, autoscaling disabled on resource %s", info.Resource.GetName())
		info.Evaluation.TargetReplicas = 0
		return &info.Evaluation, nil
	}

	// Prune old evaluations
	now := time.Now().UTC()
	// Cutoff is current time - stabilization window
	cutoff := now.Add(time.Duration(-s.Config.DownscaleStabilization) * time.Second)
	// Loop backwards over stabilization evaluations to prune old ones
	// Backwards loop to allow values to be removed mid-loop without breaking it
	for i := len(s.StabilizationEvaluations) - 1; i >= 0; i-- {
		timestampedEvaluation := s.StabilizationEvaluations[i]
		if timestampedEvaluation.Time.Before(cutoff) {
			s.StabilizationEvaluations = append(s.StabilizationEvaluations[:i], s.StabilizationEvaluations[i+1:]...)
		}
	}

	// Add to stabilization evaluations
	s.StabilizationEvaluations = append(s.StabilizationEvaluations, TimestampedEvaluation{
		Time: time.Now(),
		Evaluation: evaluate.Evaluation{
			TargetReplicas: targetReplicas,
		},
	})

	// Pick max evaluation
	for _, timestampedEvaluation := range s.StabilizationEvaluations {
		if timestampedEvaluation.Evaluation.TargetReplicas > targetReplicas {
			targetReplicas = timestampedEvaluation.Evaluation.TargetReplicas
		}
	}
	glog.V(0).Infof("Picked max evaluation over stabilization window of %d seconds; replicas %d", s.Config.DownscaleStabilization, targetReplicas)

	info.Evaluation.TargetReplicas = targetReplicas

	// Convert scaling hook input to JSON
	specJSON, err := json.Marshal(info)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	if s.Config.PreScale != nil {
		glog.V(3).Infoln("Attempting to run pre-scaling hook")
		hookResult, err := s.Execute.ExecuteWithValue(s.Config.PreScale, string(specJSON))
		if err != nil {
			return nil, fmt.Errorf("failed run pre-scaling hook: %w", err)
		}
		glog.V(3).Infof("Pre-scaling hook response: %+v", hookResult)
	}

	// Check if evaluation requires an action
	// If the resource needs scaled up/down
	if targetReplicas != currentReplicas {
		glog.V(0).Infof("Rescaling from %d to %d replicas", currentReplicas, targetReplicas)
		glog.V(3).Infoln("Attempting to parse group version")
		gvk := schema.FromAPIVersionAndKind(info.ScaleTargetRef.APIVersion, info.ScaleTargetRef.Kind)
		mapping, err := s.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse group version: %w", err)
		}
		glog.V(3).Infof("Group version parsed: %+v", mapping.Resource)

		// We are using JSON patch for a couple of reasons:
		// 1. CRD scale subresources do not support strategic patch
		// 2. Merge patching seems broken: https://github.com/kubernetes/kubernetes/issues/116311
		patch := []byte(fmt.Sprintf("[{\"op\":\"replace\",\"path\":\"/spec/replicas\",\"value\":%d}]", targetReplicas))

		glog.V(3).Infoln("Attempting to apply scaling changes to resource")

		glog.V(3).Infof("Applying patch: %s to resource %s in namespace %s", string(patch), scaleResource.Name, info.Namespace)

		_, err = s.Scaler.Scales(info.Namespace).Patch(context.Background(), mapping.Resource, scaleResource.Name, types.JSONPatchType, patch, metav1.PatchOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to apply scaling changes to resource: %w", err)
		}
		glog.V(3).Infoln("Application of scale successful")
	} else {
		glog.V(0).Infof("No change in target replicas, maintaining %d replicas", currentReplicas)
	}

	if s.Config.PostScale != nil {
		glog.V(3).Infoln("Attempting to run post-scaling hook")
		hookResult, err := s.Execute.ExecuteWithValue(s.Config.PostScale, string(specJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to run post-scaling hook: %w", err)
		}
		glog.V(3).Infof("Post-scaling hook response: %+v", hookResult)
	}

	return &info.Evaluation, nil
}

// GetScaleSubResource returns the scale subresource from the K8s scale API
func (s *Scale) GetScaleSubResource(apiVersion string, kind string, namespace string, name string) (*autoscalingv1.Scale, error) {
	glog.V(3).Infoln("Attempting to get scale subresource for managed resource")

	resourceGV, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, err
	}
	glog.V(3).Infof("Group version parsed: %+v", resourceGV)

	targetGR := schema.GroupResource{
		Group:    resourceGV.Group,
		Resource: kind,
	}

	scale, err := s.Scaler.Scales(namespace).Get(context.Background(), targetGR, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get scale subresource for resource: %w", err)
	}

	return scale, nil
}
