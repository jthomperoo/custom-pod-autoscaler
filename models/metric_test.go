package models_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/models"
)

const (
	testMetricPodName = "metric test pod"
	testMetricValue   = "metric value"
)

func getTestMetricValues() []*models.MetricValue {
	return []*models.MetricValue{
		&models.MetricValue{
			Pod:   testMetricPodName,
			Value: testMetricValue,
		},
	}
}

func TestMetric_CreateJSONWithDeployment(t *testing.T) {
	testMetric := models.Metric{
		DeploymentName: testDeploymentName,
		Metrics:        getTestMetricValues(),
		Deployment:     getTestDeployment(),
	}

	// Convert into JSON
	jsonEvaluation, err := json.Marshal(testMetric)
	if err != nil {
		t.Error(err)
	}

	// Convert JSON bytes into a JSON map to compare with test struct
	var jsonInterface interface{}
	err = json.Unmarshal(jsonEvaluation, &jsonInterface)
	if err != nil {
		t.Error(err)
	}
	jsonMap := jsonInterface.(map[string]interface{})

	// Check that deployment value is the string, not the test deployment
	deploymentName := jsonMap["deployment"].(string)
	if deploymentName != testMetric.DeploymentName {
		t.Errorf("Deployment mismatch (-want +got):\n%s", cmp.Diff(testMetric.DeploymentName, deploymentName))
	}
}
