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

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/models"
	"github.com/jthomperoo/custom-pod-autoscaler/shell"
)

// GetEvaluations uses the metrics provided to determine a set of evaluations
func GetEvaluations(metrics []*models.Metric, config *config.Config) ([]*models.Evaluation, error) {
	var evaluations []*models.Evaluation
	for _, metric := range metrics {
		evaluation, err := getEvaluationForMetric(config.Evaluate, metric, config.EvaluateTimeout)
		if err != nil {
			return nil, err
		}
		evaluations = append(evaluations, evaluation)
	}
	return evaluations, nil
}

// getEvaluationForMetric uses a metric to evaluate how to scale
func getEvaluationForMetric(cmd string, metric *models.Metric, timeout int) (*models.Evaluation, error) {
	// Convert metric into JSON
	metricJSON, err := json.Marshal(metric.Metrics)
	if err != nil {
		return nil, err
	}

	// Execute the Evaluate command with the metric JSON
	outb, err := shell.ExecWithValuePipe(cmd, string(metricJSON), timeout)
	if err != nil {
		return nil, err
	}
	evaluation := &models.EvaluationValue{}
	json.Unmarshal(outb.Bytes(), evaluation)
	return &models.Evaluation{
		DeploymentName: metric.DeploymentName,
		Evaluation:     evaluation,
		Deployment:     metric.Deployment,
	}, nil
}
