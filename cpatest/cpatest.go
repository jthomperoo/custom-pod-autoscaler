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

// Package cpatest contains utility testing methods, used in multiple tests
package cpatest

import (
	"bytes"
	"errors"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testNamespace                = "test namespace"
	testExecuteError             = "test error"
	testEvaluate                 = "test evaluate"
	testMetric                   = "test metric"
	testInterval                 = 1234
	testHost                     = "1.2.3.4"
	testPort                     = 1234
	testMetricTimeout            = 4321
	testEvaluateTimeout          = 8765
	testScaleTargetRefKind       = "test kind"
	testScaleTargetRefName       = "test name"
	testScaleTargetRefAPIVersion = "test api version"
	testExecuteSuccess           = "test success"
)

// FailExecute allows creating a shell command that will fail, i.e return an error
type FailExecute struct{}

// ExecuteWithPipe is the implementation of ExecuteWithPipe that will return an error
func (e *FailExecute) ExecuteWithPipe(command string, value string, timeout int) (*bytes.Buffer, error) {
	return nil, GetFailExecuteErr()
}

// EquateErrors creates a comparison option for cmp functions, allowing comparison of errors
func EquateErrors() cmp.Option {
	return cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})
}

// Deployment creates a deployment with test attributes
func Deployment(name, namespace string, labels map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

// GetTestConfig creates a config with test attributes
func GetTestConfig() *config.Config {
	return &config.Config{
		Evaluate:        testEvaluate,
		Metric:          testMetric,
		Interval:        testInterval,
		Host:            testHost,
		Port:            testPort,
		MetricTimeout:   testMetricTimeout,
		EvaluateTimeout: testEvaluateTimeout,
		Namespace:       testNamespace,
		ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
			Name:       testScaleTargetRefName,
			Kind:       testScaleTargetRefKind,
			APIVersion: testScaleTargetRefAPIVersion,
		},
	}
}

// GetFailExecuteErr returns the error created by the fake shell command executing FailExecute
func GetFailExecuteErr() error {
	return errors.New(testExecuteError)
}
