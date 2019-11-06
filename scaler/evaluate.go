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
	"log"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/models"
	"github.com/jthomperoo/custom-pod-autoscaler/shell"
)

// GetEvaluation uses the metrics provided to determine a set of evaluations
func GetEvaluation(resourceMetric *models.ResourceMetrics, config *config.Config, executer shell.ExecuteWithPiper) (*models.Evaluation, error) {
	// Convert metrics into JSON
	metricJSON, err := json.Marshal(resourceMetric.Metrics)
	if err != nil {
		return nil, err
	}

	// Execute the Evaluate command with the metric JSON
	outb, err := executer.ExecuteWithPipe(config.Evaluate, string(metricJSON), config.EvaluateTimeout)
	if err != nil {
		log.Println(outb.String())
		return nil, err
	}
	evaluation := &models.EvaluationValue{}
	json.Unmarshal(outb.Bytes(), evaluation)
	return &models.Evaluation{
		DeploymentName: resourceMetric.DeploymentName,
		Evaluation:     evaluation,
		Deployment:     resourceMetric.Deployment,
	}, nil
}
