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

// Package metricget provides functionality for managing gathering metrics, calling external metric gathering logic
// through shell commands with relevant data piped to them.
package metricget

import (
	"context"
	"encoding/json"
	"fmt"

	argov1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/golang/glog"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/execute"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/metric"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// GetMetricer provides methods for retrieving metrics
type GetMetricer interface {
	GetMetrics(info metric.Info) ([]*metric.ResourceMetric, error)
}

// Gatherer handles triggering the metric gathering logic to gather metrics for a resource
type Gatherer struct {
	Clientset         kubernetes.Interface
	Config            *config.Config
	Execute           execute.Executer
	K8sMetricGatherer k8smetricget.Gatherer
}

// GetMetrics gathers metrics for the resource supplied
func (m *Gatherer) GetMetrics(info metric.Info) ([]*metric.ResourceMetric, error) {

	// Query metrics server if requested
	if m.Config.KubernetesMetricSpecs != nil {
		glog.V(3).Infoln("K8s Metrics specs provided, attempting to query the K8s metrics server")
		k8sMetricSpec, err := m.K8sMetricGatherer.GetMetrics(info.Resource, m.Config.KubernetesMetricSpecs, m.Config.Namespace)
		if err != nil {
			if m.Config.RequireKubernetesMetrics {
				return nil, err
			}
			glog.Errorf("Failed to retrieve K8s metrics, not required so continuing: %+v", err)
		} else {
			glog.V(3).Infof("Successfully retrieved K8s metrics: %+v", k8sMetricSpec)
		}
		info.KubernetesMetrics = k8sMetricSpec
		glog.V(3).Infoln("Finished querying the K8s metrics server")
	}

	switch m.Config.RunMode {
	case config.PerPodRunMode:
		return m.getMetricsForPods(info)
	case config.PerResourceRunMode:
		return m.getMetricsForResource(info)
	default:
		return nil, fmt.Errorf("Unknown run mode: %s", m.Config.RunMode)
	}
}

func (m *Gatherer) getMetricsForResource(info metric.Info) ([]*metric.ResourceMetric, error) {
	glog.V(3).Infoln("Gathering metrics in per-resource mode")

	// Convert the Resource description to JSON
	specJSON, err := json.Marshal(info)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	if m.Config.PreMetric != nil {
		glog.V(3).Infoln("Attempting to run pre-metric hook")
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PreMetric, string(specJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Pre-metric hook response: %+v", hookResult)
	}

	glog.V(3).Infoln("Attempting to run metric gathering logic")
	gathered, err := m.Execute.ExecuteWithValue(m.Config.Metric, string(specJSON))
	if err != nil {
		return nil, err
	}
	info.Metrics = []*metric.ResourceMetric{
		{
			Resource: info.Resource.GetName(),
			Value:    gathered,
		},
	}
	glog.V(3).Infof("Metrics gathered: %+v", gathered)

	if m.Config.PostMetric != nil {
		glog.V(3).Infoln("Attempting to run post-metric hook")
		// Convert post metrics into JSON
		postSpecJSON, err := json.Marshal(info)
		if err != nil {
			// Should not occur, panic
			panic(err)
		}
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PostMetric, string(postSpecJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Post-metric hook response: %+v", hookResult)
	}

	return info.Metrics, nil
}

func (m *Gatherer) getMetricsForPods(info metric.Info) ([]*metric.ResourceMetric, error) {
	glog.V(3).Infoln("Gathering metrics in per-pod mode")

	glog.V(3).Infoln("Attempting to get pod selector from managed resource")
	labels, err := m.getPodSelectorForResource(info.Resource)
	if err != nil {
		return nil, err
	}
	glog.V(3).Infof("Label selector retrieved: %+v", labels)

	glog.V(3).Infoln("Attempting to get pods being managed")
	pods, err := m.Clientset.CoreV1().Pods(m.Config.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: labels})
	if err != nil {
		return nil, err
	}
	glog.V(3).Infof("Pods retrieved: %+v", pods)

	// Convert the metric info into JSON
	specJSON, err := json.Marshal(info)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	if m.Config.PreMetric != nil {
		glog.V(3).Infoln("Attempting to run pre-metric hook")
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PreMetric, string(specJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Pre-metric hook response: %+v", hookResult)
	}

	glog.V(3).Infoln("Attempting to gather metrics for each pod")
	var metrics []*metric.ResourceMetric
	for _, pod := range pods.Items {
		// Convert the Pod description to JSON
		podSpecJSON, err := json.Marshal(metric.Info{
			Resource:          &pod,
			RunType:           info.RunType,
			KubernetesMetrics: info.KubernetesMetrics,
		})
		if err != nil {
			// Should not occur, panic
			panic(err)
		}

		glog.V(3).Infof("Running metric gathering for pod: %s", pod.Name)
		gathered, err := m.Execute.ExecuteWithValue(m.Config.Metric, string(podSpecJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Metric gathered: %+v", gathered)

		// Add metric to metrics array
		metrics = append(metrics, &metric.ResourceMetric{
			Resource: pod.GetName(),
			Value:    gathered,
		})
	}
	glog.V(3).Infoln("All metrics gathered for each pod successfully")
	info.Metrics = metrics

	if m.Config.PostMetric != nil {
		glog.V(3).Infoln("Attempting to run post-metric hook")
		// Convert post metrics into JSON
		postSpecJSON, err := json.Marshal(info)
		if err != nil {
			// Should not occur, panic
			panic(err)
		}
		hookResult, err := m.Execute.ExecuteWithValue(m.Config.PostMetric, string(postSpecJSON))
		if err != nil {
			return nil, err
		}
		glog.V(3).Infof("Post-metric hook response: %+v", hookResult)
	}

	return metrics, nil
}

func (m *Gatherer) getPodSelectorForResource(resource metav1.Object) (string, error) {
	switch v := resource.(type) {
	case *appsv1.Deployment:
		selector, err := metav1.LabelSelectorAsMap(v.Spec.Selector)
		if err != nil {
			return "", err
		}
		return labels.SelectorFromSet(selector).String(), nil
	case *appsv1.ReplicaSet:
		selector, err := metav1.LabelSelectorAsMap(v.Spec.Selector)
		if err != nil {
			return "", err
		}
		return labels.SelectorFromSet(selector).String(), nil
	case *appsv1.StatefulSet:
		selector, err := metav1.LabelSelectorAsMap(v.Spec.Selector)
		if err != nil {
			return "", err
		}
		return labels.SelectorFromSet(selector).String(), nil
	case *corev1.ReplicationController:
		return labels.SelectorFromSet(v.Spec.Selector).String(), nil
	case *argov1alpha1.Rollout:
		selector, err := metav1.LabelSelectorAsMap(v.Spec.Selector)
		if err != nil {
			return "", err
		}
		return labels.SelectorFromSet(selector).String(), nil
	default:
		return "", fmt.Errorf("Unsupported resource of type %T", v)
	}
}
