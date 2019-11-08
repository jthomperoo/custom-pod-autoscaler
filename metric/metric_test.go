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
// +build unit

package metric_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/cpatest"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/models"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testManagedDeploymentName     = "test managed deployment"
	testManagedNamespace          = "test managed namespace"
	testManagedDeploymentAppLabel = "test-managed"

	testUnmanagedDeploymentName     = "test unmanaged deployment"
	testUnmanagedNamespace          = "test unmanaged namespace"
	testUnmanagedDeploymentAppLabel = "test-unmanaged"

	testFirstPodName  = "first pod"
	testSecondPodName = "second pod"

	testMetricValue = "test value"
)

type executeWithPiper interface {
	ExecuteWithPipe(command string, value string, timeout int) (*bytes.Buffer, error)
}

type successExecuteMetric struct{}

func (e *successExecuteMetric) ExecuteWithPipe(command string, value string, timeout int) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	buffer.WriteString(testMetricValue)
	return &buffer, nil
}

func TestGetMetrics(t *testing.T) {
	var tests = []struct {
		description string
		expectedErr error
		expected    []*models.Metric
		deployment  *appsv1.Deployment
		config      *config.Config
		clientset   kubernetes.Interface
		executer    executeWithPiper
	}{
		{
			"No resources",
			nil,
			nil,
			&appsv1.Deployment{},
			getConfig(testManagedNamespace),
			fake.NewSimpleClientset(),
			&successExecuteMetric{},
		},
		{
			"No pod in managed deployment, but pod in other deployment with different name in same namespace",
			nil,
			nil,
			cpatest.Deployment(testManagedDeploymentName, testManagedNamespace, map[string]string{"app": testManagedDeploymentAppLabel}),
			getConfig(testManagedNamespace),
			fake.NewSimpleClientset(
				pod(testFirstPodName, testManagedNamespace, map[string]string{"app": testUnmanagedDeploymentAppLabel}),
			),
			&successExecuteMetric{},
		},
		{
			"No pod in managed deployment, but pod in other deployment with same name in different namespace",
			nil,
			nil,
			cpatest.Deployment(testManagedDeploymentName, testManagedNamespace, map[string]string{"app": testManagedDeploymentAppLabel}),
			getConfig(testManagedNamespace),
			fake.NewSimpleClientset(
				pod(testFirstPodName, testUnmanagedNamespace, map[string]string{"app": testManagedDeploymentAppLabel}),
			),
			&successExecuteMetric{},
		},
		{
			"Single pod single deployment shell execute fail",
			cpatest.GetFailExecuteErr(),
			nil,
			cpatest.Deployment(testManagedDeploymentName, testManagedNamespace, map[string]string{"app": testManagedDeploymentAppLabel}),
			getConfig(testManagedNamespace),
			fake.NewSimpleClientset(
				pod(testFirstPodName, testManagedNamespace, map[string]string{"app": testManagedDeploymentAppLabel}),
			),
			&cpatest.FailExecute{},
		},
		{
			"Single pod single deployment shell execute success",
			nil,
			[]*models.Metric{
				getTestMetric(testFirstPodName, testMetricValue),
			},
			cpatest.Deployment(testManagedDeploymentName, testManagedNamespace, map[string]string{"app": testManagedDeploymentAppLabel}),
			getConfig(testManagedNamespace),
			fake.NewSimpleClientset(
				pod(testFirstPodName, testManagedNamespace, map[string]string{"app": testManagedDeploymentAppLabel}),
			),
			&successExecuteMetric{},
		},
		{
			"Multiple pod single deployment shell execute success",
			nil,
			[]*models.Metric{
				getTestMetric(testFirstPodName, testMetricValue),
				getTestMetric(testSecondPodName, testMetricValue),
			},
			cpatest.Deployment(testManagedDeploymentName, testManagedNamespace, map[string]string{"app": testManagedDeploymentAppLabel}),
			getConfig(testManagedNamespace),
			fake.NewSimpleClientset(
				pod(testFirstPodName, testManagedNamespace, map[string]string{"app": testManagedDeploymentAppLabel}),
				pod(testSecondPodName, testManagedNamespace, map[string]string{"app": testManagedDeploymentAppLabel}),
			),
			&successExecuteMetric{},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := &models.ResourceMetrics{
				Metrics:        test.expected,
				Deployment:     test.deployment,
				DeploymentName: test.deployment.Name,
			}
			gatherer := &metric.Gatherer{
				Clientset: test.clientset,
				Config:    test.config,
				Executer:  test.executer,
			}
			metrics, err := gatherer.GetMetrics(test.deployment)
			if !cmp.Equal(err, test.expectedErr, cpatest.EquateErrors()) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, cpatest.EquateErrors()))
				return
			}
			if test.expectedErr != nil {
				return
			}
			if !cmp.Equal(metrics, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(result, metrics))
			}
		})
	}
}

func pod(name, namespace string, labels map[string]string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
	}
}

func getConfig(namespace string) *config.Config {
	return &config.Config{
		Namespace: namespace,
	}
}

func getTestResourceMetrics(metrics []*models.Metric, deployment *appsv1.Deployment) *models.ResourceMetrics {
	return &models.ResourceMetrics{
		Metrics:        metrics,
		DeploymentName: deployment.Name,
		Deployment:     deployment,
	}
}

func getTestMetric(pod, value string) *models.Metric {
	return &models.Metric{
		Pod:   pod,
		Value: value,
	}
}
