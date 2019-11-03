package models_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/models"
)

func getTestEvaluationValue() *models.EvaluationValue {
	return &models.EvaluationValue{
		TargetReplicas: testEvaluationValueTargetReplicas,
	}
}

func TestEvaluation_CreateJSONWithDeployment(t *testing.T) {
	testEvaluation := models.Evaluation{
		DeploymentName: testDeploymentName,
		Evaluation:     getTestEvaluationValue(),
		Deployment:     getTestDeployment(),
	}

	// Convert into JSON
	jsonEvaluation, err := json.Marshal(testEvaluation)
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
	if deploymentName != testEvaluation.DeploymentName {
		t.Errorf("Deployment mismatch (-want +got):\n%s", cmp.Diff(testEvaluation.DeploymentName, deploymentName))
	}
}
