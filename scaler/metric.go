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

package scaler

import (
	"encoding/json"
	"fmt"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/models"
	"github.com/jthomperoo/custom-pod-autoscaler/shell"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetMetrics gathers metrics for the deployments supplied
func GetMetrics(clientset *kubernetes.Clientset, deployments *appsv1.DeploymentList, config *config.Config) ([]*models.Metric, error) {
	var metrics []*models.Metric
	for _, deployment := range deployments.Items {
		// Get Deployment pods
		labels := deployment.GetLabels()
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", labels["app"])})
		if err != nil {
			return nil, err
		}

		// Gather metrics for each pod
		var metricValues []*models.MetricValue
		for _, pod := range pods.Items {
			metric, err := getMetricForPod(config.Metric, &pod)
			if err != nil {
				continue
			}
			metricValues = append(metricValues, metric)
		}

		metrics = append(metrics, &models.Metric{
			DeploymentName: deployment.GetName(),
			Deployment:     &deployment,
			Metrics:        metricValues,
		})
	}
	return metrics, nil
}

// getMetricForPod gathers the metric for a specific pod
func getMetricForPod(metricCmd string, pod *corev1.Pod) (*models.MetricValue, error) {
	// Convert the Pod description to JSON
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}

	// Execute the Metric command with the Pod JSON
	outb, err := shell.ExecWithValuePipe(metricCmd, string(podJSON))
	if err != nil {
		return nil, err
	}

	return &models.MetricValue{
		Pod:   pod.GetName(),
		Value: string(outb.String()),
	}, nil
}
