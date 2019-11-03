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
