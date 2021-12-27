/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Modifications Copyright 2021 The Custom Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.

Modified to split up evaluations and metric gathering to work with the
Custom Pod Autoscaler framework.
Original source:
https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/podautoscaler/horizontal.go
https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/podautoscaler/replica_calculator.go
*/

// Package k8smetricget provides K8s metric gathering, in the same way that the Horizontal Pod Autoscaler gathers
// metrics, using the metrics APIs.
package k8smetricget

import (
	"fmt"
	"time"

	argov1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/externalget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/objectget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/podsget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/podutil"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/resourceget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corelisters "k8s.io/client-go/listers/core/v1"
	metricsclient "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

// Gatherer allows retrieval of metrics.
type Gatherer interface {
	GetMetrics(resource metav1.Object, specs []config.K8sMetricSpec, namespace string) ([]*k8smetric.Metric, error)
}

// Gather provides functionality for retrieving metrics on supplied metric specs.
type Gather struct {
	Resource resourceget.Gatherer
	Pods     podsget.Gatherer
	Object   objectget.Gatherer
	External externalget.Gatherer
}

// NewGather sets up a new Metric Gatherer
func NewGather(
	metricsClient metricsclient.MetricsClient,
	podlister corelisters.PodLister,
	cpuInitializationPeriod time.Duration,
	delayOfInitialReadinessStatus time.Duration) *Gather {

	// Set up pod ready counter
	podReadyCounter := &podutil.PodReadyCount{
		PodLister: podlister,
	}

	return &Gather{
		Resource: &resourceget.Gather{
			MetricsClient:                 metricsClient,
			PodLister:                     podlister,
			CPUInitializationPeriod:       cpuInitializationPeriod,
			DelayOfInitialReadinessStatus: delayOfInitialReadinessStatus,
		},
		Pods: &podsget.Gather{
			MetricsClient: metricsClient,
			PodLister:     podlister,
		},
		Object: &objectget.Gather{
			MetricsClient:   metricsClient,
			PodReadyCounter: podReadyCounter,
		},
		External: &externalget.Gather{
			MetricsClient:   metricsClient,
			PodReadyCounter: podReadyCounter,
		},
	}
}

// GetMetrics processes each MetricSpec provided, calculating metric values for each and combining them into a slice before returning them.
// Error will only be returned if all metrics are invalid, otherwise it will return the valid metrics.
func (c *Gather) GetMetrics(resource metav1.Object, specs []config.K8sMetricSpec, namespace string) ([]*k8smetric.Metric, error) {
	var combinedMetrics []*k8smetric.Metric
	var invalidMetricError error
	invalidMetricsCount := 0
	currentReplicas := int32(0)
	resourceReplicas, err := c.getReplicaCount(resource)
	if err != nil {
		return nil, err
	}
	if resourceReplicas != nil {
		currentReplicas = *resourceReplicas
	}
	for _, spec := range specs {
		metric, err := c.getMetric(currentReplicas, spec, namespace, labels.Set(resource.GetLabels()).AsSelector())
		if err != nil {
			if invalidMetricsCount <= 0 {
				invalidMetricError = err
			}
			invalidMetricsCount++
			continue
		}
		combinedMetrics = append(combinedMetrics, metric)
	}

	// If all metrics are invalid return error and set condition on hpa based on first invalid metric.
	if invalidMetricsCount >= len(specs) {
		return nil, fmt.Errorf("invalid metrics (%v invalid out of %v), first error is: %v", invalidMetricsCount, len(specs), invalidMetricError)
	}

	return combinedMetrics, nil
}

func (c *Gather) getMetric(currentReplicas int32, spec config.K8sMetricSpec, namespace string, selector labels.Selector) (*k8smetric.Metric, error) {
	switch spec.Type {
	case autoscaling.ObjectMetricSourceType:
		metricSelector, err := metav1.LabelSelectorAsSelector(spec.Object.Metric.Selector)
		if err != nil {
			return nil, fmt.Errorf("failed to get object metric: %v", err)
		}

		switch spec.Object.Target.Type {
		case autoscaling.ValueMetricType:
			objectMetric, err := c.Object.GetMetric(spec.Object.Metric.Name, namespace, &spec.Object.DescribedObject, selector, metricSelector)
			if err != nil {
				return nil, fmt.Errorf("failed to get object metric: %v", err)
			}
			return &k8smetric.Metric{
				CurrentReplicas: currentReplicas,
				Spec:            spec,
				Object:          objectMetric,
			}, nil
		case autoscaling.AverageValueMetricType:
			objectMetric, err := c.Object.GetPerPodMetric(spec.Object.Metric.Name, namespace, &spec.Object.DescribedObject, selector)
			if err != nil {
				return nil, fmt.Errorf("failed to get object metric: %v", err)
			}
			return &k8smetric.Metric{
				CurrentReplicas: currentReplicas,
				Spec:            spec,
				Object:          objectMetric,
			}, nil
		default:
			return nil, fmt.Errorf("invalid object metric source: must be either value or average value")
		}
	case autoscaling.PodsMetricSourceType:
		metricSelector, err := metav1.LabelSelectorAsSelector(spec.Pods.Metric.Selector)
		if err != nil {
			return nil, fmt.Errorf("failed to get pods metric: %v", err)
		}

		if spec.Pods.Target.Type != autoscaling.AverageValueMetricType {
			return nil, fmt.Errorf("invalid pods metric source: must be average value")
		}

		podsMetric, err := c.Pods.GetMetric(spec.Pods.Metric.Name, namespace, selector, metricSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to get pods metric: %v", err)
		}
		return &k8smetric.Metric{
			CurrentReplicas: currentReplicas,
			Spec:            spec,
			Pods:            podsMetric,
		}, nil
	case autoscaling.ResourceMetricSourceType:
		switch spec.Resource.Target.Type {
		case autoscaling.AverageValueMetricType:
			resourceMetric, err := c.Resource.GetRawMetric(spec.Resource.Name, namespace, selector)
			if err != nil {
				return nil, fmt.Errorf("failed to get resource metric: %v", err)
			}
			return &k8smetric.Metric{
				CurrentReplicas: currentReplicas,
				Spec:            spec,
				Resource:        resourceMetric,
			}, nil
		case autoscaling.UtilizationMetricType:
			resourceMetric, err := c.Resource.GetMetric(spec.Resource.Name, namespace, selector)
			if err != nil {
				return nil, fmt.Errorf("failed to get resource metric: %v", err)
			}
			return &k8smetric.Metric{
				CurrentReplicas: currentReplicas,
				Spec:            spec,
				Resource:        resourceMetric,
			}, nil
		default:
			return nil, fmt.Errorf("invalid resource metric source: must be either average value or average utilization")
		}

	case autoscaling.ExternalMetricSourceType:
		switch spec.External.Target.Type {
		case autoscaling.AverageValueMetricType:
			externalMetric, err := c.External.GetPerPodMetric(spec.External.Metric.Name, namespace, spec.External.Metric.Selector)
			if err != nil {
				return nil, fmt.Errorf("failed to get external metric: %v", err)
			}
			return &k8smetric.Metric{
				CurrentReplicas: currentReplicas,
				Spec:            spec,
				External:        externalMetric,
			}, nil
		case autoscaling.ValueMetricType:
			externalMetric, err := c.External.GetMetric(spec.External.Metric.Name, namespace, spec.External.Metric.Selector, selector)
			if err != nil {
				return nil, fmt.Errorf("failed to get external metric: %v", err)
			}
			return &k8smetric.Metric{
				CurrentReplicas: currentReplicas,
				Spec:            spec,
				External:        externalMetric,
			}, nil
		default:
			return nil, fmt.Errorf("invalid external metric source: must be either value or average value")
		}

	default:
		return nil, fmt.Errorf("unknown metric source type %q", string(spec.Type))
	}
}

func (c *Gather) getReplicaCount(resource metav1.Object) (*int32, error) {
	switch v := resource.(type) {
	case *appsv1.Deployment:
		return v.Spec.Replicas, nil
	case *appsv1.ReplicaSet:
		return v.Spec.Replicas, nil
	case *appsv1.StatefulSet:
		return v.Spec.Replicas, nil
	case *corev1.ReplicationController:
		return v.Spec.Replicas, nil
	case *argov1alpha1.Rollout:
		return v.Spec.Replicas, nil
	default:
		return nil, fmt.Errorf("Unsupported resource of type %T", v)
	}
}
