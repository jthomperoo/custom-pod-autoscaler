package scaler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/models"
	"github.com/jthomperoo/custom-pod-autoscaler/scaler"
)

const (
	testMetricPod                   = "test pod"
	testMetricValue                 = "test value"
	testEvaluationInvalidEvaluation = "{ \"invalid\": \"invalid\"}"
	testEvaluationInvalidJSON       = "invalid}"
	testEvaluationTargetReplicas    = int32(3)
)

type successExecuteValidEvaluation struct{}

func (e *successExecuteValidEvaluation) ExecuteWithPipe(command string, value string, timeout int) (*bytes.Buffer, error) {
	// Convert into JSON
	jsonEvaluation, err := json.Marshal(getTestEvaluation())
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	buffer.WriteString(string(jsonEvaluation))
	return &buffer, nil
}

type successExecuteInvalidEvaluation struct{}

func (e *successExecuteInvalidEvaluation) ExecuteWithPipe(command string, value string, timeout int) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	buffer.WriteString(testEvaluationInvalidEvaluation)
	return &buffer, nil
}

type successExecuteInvalidJSON struct{}

func (e *successExecuteInvalidJSON) ExecuteWithPipe(command string, value string, timeout int) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	buffer.WriteString(testEvaluationInvalidJSON)
	return &buffer, nil
}

func TestGetEvaluation_ExecuteFail(t *testing.T) {
	resourceMetrics := getTestResourceMetrics()
	config := getTestConfig()

	_, err := scaler.GetEvaluation(resourceMetrics, config, &failExecuteWithPipe{})
	if err == nil {
		t.Errorf("Expected error due to executer failing and returning an error")
		return
	}

	if err.Error() != testExecuteError {
		t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(testExecuteError, err.Error()))
	}
}

func TestGetEvaluation_ExecuteSuccessValidJSON(t *testing.T) {
	resourceMetrics := getTestResourceMetrics()
	config := getTestConfig()
	testEvaluation := getTestEvaluation()

	evaluation, err := scaler.GetEvaluation(resourceMetrics, config, &successExecuteValidEvaluation{})
	if err != nil {
		t.Error(err)
		return
	}

	if !cmp.Equal(testEvaluation, evaluation) {
		t.Errorf("Evaluation mismatch (-want +got):\n%s", cmp.Diff(testEvaluation, evaluation))
	}
}

func TestGetEvaluation_ExecuteSuccessInvalidEvaluation(t *testing.T) {
	resourceMetrics := getTestResourceMetrics()
	config := getTestConfig()
	_, err := scaler.GetEvaluation(resourceMetrics, config, &successExecuteInvalidEvaluation{})

	if err == nil {
		t.Errorf("Expected error due to executer returning an invalid evaluation")
		return
	}

	if _, ok := err.(*scaler.ErrInvalidEvaluation); !ok {
		t.Errorf("Expected invalid evaluation, instead got: %v", err)
	}

	if err.Error() != fmt.Sprintf("Invalid evaluation returned by evaluator: %s", testEvaluationInvalidEvaluation) {
		t.Errorf("Error mismatch (-want +got):\n%s", cmp.Diff(testEvaluationInvalidEvaluation, err.Error()))
	}
}

func TestGetEvaluation_ExecuteSuccessInvalidJSONSyntax(t *testing.T) {
	resourceMetrics := getTestResourceMetrics()
	config := getTestConfig()
	_, err := scaler.GetEvaluation(resourceMetrics, config, &successExecuteInvalidJSON{})

	if err == nil {
		t.Errorf("Expected error due to executer returning invalid JSON syntax")
		return
	}

	if _, ok := err.(*json.SyntaxError); !ok {
		t.Errorf("Expected invalid JSON syntax, instead got: %v", err)
	}
}

func getTestResourceMetrics() *models.ResourceMetrics {
	return &models.ResourceMetrics{
		DeploymentName: testDeploymentName,
		Deployment:     getTestDeployment(),
		Metrics: []*models.Metric{
			&models.Metric{
				Pod:   testMetricPod,
				Value: testMetricValue,
			},
		},
	}
}

func getTestEvaluation() *models.Evaluation {
	targetReplicas := testEvaluationTargetReplicas
	return &models.Evaluation{
		TargetReplicas: &targetReplicas,
	}
}
