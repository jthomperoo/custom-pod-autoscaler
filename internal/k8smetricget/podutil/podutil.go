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

Modifications Copyright 2022 The Custom Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.

Modified to split up evaluations and metric gathering to work with the
Custom Pod Autoscaler framework.
Original source:
https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/podautoscaler/horizontal.go
https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/podautoscaler/replica_calculator.go
*/

package podutil

import (
	"fmt"
	"time"

	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/podmetrics"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// PodReadyCounter provides a way to count number of ready pods
type PodReadyCounter interface {
	GetReadyPodsCount(namespace string, selector labels.Selector) (int64, error)
}

// PodReadyCount provides a way to count the number of ready pods using a pod lister
type PodReadyCount struct {
	PodLister corelisters.PodLister
}

// GetReadyPodsCount returns the number of pods that are deemed 'ready'
func (c *PodReadyCount) GetReadyPodsCount(namespace string, selector labels.Selector) (int64, error) {
	// Get pods
	podList, err := c.PodLister.Pods(namespace).List(selector)
	if err != nil {
		return 0, fmt.Errorf("unable to get pods while calculating replica count: %w", err)
	}

	// Count number of ready pods
	readyPodCount := int64(0)
	for _, pod := range podList {
		if pod.Status.Phase == corev1.PodRunning && isPodReady(pod) {
			readyPodCount++
		}
	}

	return readyPodCount, nil
}

// GroupPods groups pods into ready, missing and ignored based on PodMetricsInfo and resource provided
func GroupPods(pods []*corev1.Pod, metrics podmetrics.MetricsInfo, resource corev1.ResourceName, cpuInitializationPeriod, delayOfInitialReadinessStatus time.Duration) (readyPodCount int, ignoredPods sets.String, missingPods sets.String) {
	missingPods = sets.NewString()
	ignoredPods = sets.NewString()
	for _, pod := range pods {
		if pod.DeletionTimestamp != nil || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		// Pending pods are ignored.
		if pod.Status.Phase == corev1.PodPending {
			ignoredPods.Insert(pod.Name)
			continue
		}
		// Pods missing metrics.
		metric, found := metrics[pod.Name]
		if !found {
			missingPods.Insert(pod.Name)
			continue
		}
		// Unready pods are ignored.
		if resource == corev1.ResourceCPU {
			var ignorePod bool
			_, condition := getPodCondition(pod.Status, corev1.PodReady)
			if condition == nil || pod.Status.StartTime == nil {
				ignorePod = true
			} else {
				// Pod still within possible initialisation period.
				if pod.Status.StartTime.Add(cpuInitializationPeriod).After(time.Now()) {
					// Ignore sample if pod is unready or one window of metric wasn't collected since last state transition.
					ignorePod = condition.Status == corev1.ConditionFalse || metric.Timestamp.Before(condition.LastTransitionTime.Time.Add(metric.Window))
				} else {
					// Ignore metric if pod is unready and it has never been ready.
					ignorePod = condition.Status == corev1.ConditionFalse && pod.Status.StartTime.Add(delayOfInitialReadinessStatus).After(condition.LastTransitionTime.Time)
				}
			}
			if ignorePod {
				ignoredPods.Insert(pod.Name)
				continue
			}
		}
		readyPodCount++
	}
	return
}

// CalculatePodRequests calculates pod resource requests for a slice of pods
func CalculatePodRequests(pods []*corev1.Pod, resource corev1.ResourceName) (map[string]int64, error) {
	requests := make(map[string]int64, len(pods))
	for _, pod := range pods {
		podSum := int64(0)
		for _, container := range pod.Spec.Containers {
			if containerRequest, ok := container.Resources.Requests[resource]; ok {
				podSum += containerRequest.MilliValue()
			} else {
				return nil, fmt.Errorf("missing request for %s", resource)
			}
		}
		requests[pod.Name] = podSum
	}
	return requests, nil
}

// RemoveMetricsForPods removes the pods provided from the PodMetricsInfo provided
func RemoveMetricsForPods(metrics podmetrics.MetricsInfo, pods sets.String) {
	for _, pod := range pods.UnsortedList() {
		delete(metrics, pod)
	}
}

// IsPodReady returns true if a pod is ready; false otherwise.
func isPodReady(pod *corev1.Pod) bool {
	_, condition := getPodCondition(pod.Status, corev1.PodReady)
	return condition != nil && condition.Status == corev1.ConditionTrue
}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func getPodCondition(status corev1.PodStatus, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	conditions := status.Conditions
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}
