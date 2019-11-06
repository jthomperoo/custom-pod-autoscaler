package scaler_test

import (
	"bytes"
	"errors"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	invalidYAML                  = "- in: -: valid - yaml"
	testEvaluate                 = "test evaluate"
	testMetric                   = "test metric"
	testInterval                 = 1234
	testHost                     = "1.2.3.4"
	testPort                     = 1234
	testMetricTimeout            = 4321
	testEvaluateTimeout          = 8765
	testNamespace                = "test namespace"
	testScaleTargetRefKind       = "test kind"
	testScaleTargetRefName       = "test name"
	testScaleTargetRefAPIVersion = "test api version"
	testDeploymentName           = "test deployment"
	testExecuteError             = "test error"
	testExecuteSuccess           = "test success"
)

type failExecuteWithPipe struct{}

func (e *failExecuteWithPipe) ExecuteWithPipe(command string, value string, timeout int) (*bytes.Buffer, error) {
	return nil, errors.New(testExecuteError)
}

func getTestDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: testDeploymentName,
		},
	}
}

func getTestConfig() *config.Config {
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
