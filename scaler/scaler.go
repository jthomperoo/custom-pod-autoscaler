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
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/shell"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

// ConfigureScaler sets up the scaler logic, which will repeatedly determine through gathering metrics and
// evaluating the metrics if the managed deployments need scaled up/down
func ConfigureScaler(clientset *kubernetes.Clientset, deploymentsClient v1.DeploymentInterface, config *config.Config, executer shell.Executer) {
	ticker := time.NewTicker(time.Duration(config.Interval) * time.Millisecond)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go scale(clientset, deploymentsClient, config, ticker, shutdown, executer)
}

func scale(clientset *kubernetes.Clientset, deploymentsClient v1.DeploymentInterface, config *config.Config, ticker *time.Ticker, shutdown chan os.Signal, executer shell.Executer) {
	for {
		select {
		case <-shutdown:
			ticker.Stop()
			return
		case <-ticker.C:
			// Get deployments being managed
			deployments, err := deploymentsClient.List(metav1.ListOptions{LabelSelector: config.Selector})
			if err != nil {
				log.Fatalf(err.Error())
			}

			// Gather metrics
			metrics, err := GetMetrics(clientset, deployments, config, executer)
			if err != nil {
				log.Printf("Failed to gather metrics\n%v", err)
				break
			}

			// Evaluate based on metrics
			evaluations, err := GetEvaluations(metrics, config, executer)
			if err != nil {
				log.Printf("Failed to evaluate metrics\n%v", err)
				break
			}

			// Check if each evaluation requires an action
			for _, evaluation := range evaluations {
				deployment := evaluation.Deployment
				// If the deployment needs scaled up/down
				if evaluation.Evaluation.TargetReplicas != *deployment.Spec.Replicas {
					// Scale deployment
					deployment.Spec.Replicas = &evaluation.Evaluation.TargetReplicas
					_, err = deploymentsClient.Update(deployment)
					if err != nil {
						log.Fatalf(err.Error())
					}
				}
			}
		}
	}
}
