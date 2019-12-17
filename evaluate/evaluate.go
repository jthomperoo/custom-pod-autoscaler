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

// Package evaluate provides functionality for managing evaluating,
// calling external evaluation logic through shell commands with
// relevant data piped to them.
package evaluate

import (
	"encoding/json"
	"log"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/execute"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
)

const invalidEvaluationMessage = "Invalid evaluation returned by evaluator: %s"

// Evaluation represents a decision on how to scale a deployment
type Evaluation struct {
	TargetReplicas int32 `json:"target_replicas"`
}

// Evaluator handles triggering the evaluation logic to decide how to scale a resource
type Evaluator struct {
	Config  *config.Config
	Execute execute.Executer
}

// GetEvaluation uses the metrics provided to determine a set of evaluations
func (e *Evaluator) GetEvaluation(resourceMetrics *metric.ResourceMetrics) (*Evaluation, error) {
	// Convert metrics into JSON
	metricJSON, err := json.Marshal(resourceMetrics)
	if err != nil {
		// Should not occur, panic
		log.Panic(err)
	}

	// Execute with the value
	gathered, err := e.Execute.ExecuteWithValue(e.Config.Evaluate, string(metricJSON))
	if err != nil {
		return nil, err
	}

	evaluation := &Evaluation{}
	err = json.Unmarshal([]byte(gathered), evaluation)
	if err != nil {
		return nil, err
	}
	return evaluation, nil
}
