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

// Package autoscaler provides methods for scaling a resource - by triggering metric
// gathering, feeding these metrics to an evaluation and using this evaluation
// to scale the resource.
package autoscaler

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/evaluatecalc"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/metricget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/resourceclient"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/scaling"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/scale"
)

// Scaler handles scaling up/down the resource being managed; triggering metric gathering and
// feeding an evaluator these metrics, before taking the results and using them to interact with Kubernetes
// to scale up/down
type Scaler struct {
	Scaler          scaling.Scaler
	Client          resourceclient.Client
	Config          *config.Config
	GetMetricer     metricget.GetMetricer
	GetEvaluationer evaluatecalc.GetEvaluationer
}

// Scale gets the managed resource, gathers metrics, evaluates these metrics and finally if a change is required
// then it scales the resource
func (s *Scaler) Scale() error {
	glog.V(2).Infoln("Attempting to get managed resource")
	resource, err := s.Client.Get(s.Config.ScaleTargetRef.APIVersion, s.Config.ScaleTargetRef.Kind, s.Config.ScaleTargetRef.Name, s.Config.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get managed resource: %w", err)
	}
	glog.V(2).Infof("Managed resource retrieved: %+v", resource)

	glog.V(2).Infoln("Attempting to get scale subresource")
	scaleResource, err := s.Scaler.GetScaleSubResource(s.Config.ScaleTargetRef.APIVersion, s.Config.ScaleTargetRef.Kind, s.Config.Namespace, s.Config.ScaleTargetRef.Name)
	if err != nil {
		return fmt.Errorf("failed to get scale subresource: %w", err)
	}
	glog.V(2).Infof("Managed scale subresource retrieved: %+v", scaleResource)

	glog.V(2).Infoln("Attempting to get resource metrics")
	metrics, err := s.GetMetricer.GetMetrics(metric.Info{
		Resource: resource,
		RunType:  config.ScalerRunType,
	}, scaleResource)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}
	glog.V(2).Infof("Retrieved metrics: %+v", metrics)

	glog.V(2).Infoln("Attempting to evaluate metrics")
	evaluation, err := s.GetEvaluationer.GetEvaluation(evaluate.Info{
		Metrics:  metrics,
		Resource: resource,
		RunType:  config.ScalerRunType,
	})
	if err != nil {
		return fmt.Errorf("failed get evaluation: %w", err)
	}
	glog.V(2).Infof("Metrics evaluated: %+v", evaluation)

	glog.V(2).Infoln("Attempting to scale resource based on evaluation")
	_, err = s.Scaler.Scale(scale.Info{
		Evaluation:     *evaluation,
		Resource:       resource,
		MinReplicas:    s.Config.MinReplicas,
		MaxReplicas:    s.Config.MaxReplicas,
		TargetReplicas: evaluation.TargetReplicas,
		Namespace:      s.Config.Namespace,
		ScaleTargetRef: s.Config.ScaleTargetRef,
		RunType:        config.ScalerRunType,
	}, scaleResource)
	if err != nil {
		return fmt.Errorf("failed to scale resource: %w", err)
	}
	glog.V(2).Infoln("Scaled resource successfully")
	return nil
}
