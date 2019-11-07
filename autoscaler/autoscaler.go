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

package autoscaler

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/models"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

type getMetricer interface {
	GetMetrics(deployment *appsv1.Deployment) (*models.ResourceMetrics, error)
}

type getEvaluationer interface {
	GetEvaluation(resourceMetrics *models.ResourceMetrics) (*models.Evaluation, error)
}

// Autoscaler handles automatically scaling up/down the resource being managed; triggering metric gathering and
// feeding an evaluator these metrics, before taking the results and using them to interact with Kubernetes
// to scale up/down
type Autoscaler struct {
	Clientset         *kubernetes.Clientset
	DeploymentsClient v1.DeploymentInterface
	Config            *config.Config
	Ticker            *time.Ticker
	Shutdown          chan os.Signal
	GetMetricer       getMetricer
	GetEvaluationer   getEvaluationer
}

// NewAutoscaler creates a new Autoscaler with some default configuration, setting up the interval
// and shutdown channel and signals
func NewAutoscaler(
	clientset *kubernetes.Clientset,
	deploymentsClient v1.DeploymentInterface,
	config *config.Config,
	getMetricer getMetricer,
	getEvaluationer getEvaluationer) *Autoscaler {
	// Set up time ticker with configured interval
	ticker := time.NewTicker(time.Duration(config.Interval) * time.Millisecond)
	// Set up shutdown channel, which will listen for UNIX shutdown commands
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	return &Autoscaler{
		Clientset:         clientset,
		DeploymentsClient: deploymentsClient,
		Config:            config,
		Ticker:            ticker,
		Shutdown:          shutdown,
		GetMetricer:       getMetricer,
		GetEvaluationer:   getEvaluationer,
	}
}

// Start kicks off the Autoscaler, which will run in a goroutine
func (a *Autoscaler) Start() {
	go a.autoscale()
}

func (a *Autoscaler) autoscale() {
	for {
		select {
		case <-a.Shutdown:
			a.Ticker.Stop()
			return
		case <-a.Ticker.C:
			// Get deployment being managed
			deployment, err := a.Clientset.AppsV1().Deployments(a.Config.Namespace).Get(a.Config.ScaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					log.Println(err)
					break
				} else {
					log.Fatal(err)
				}
			}

			// Gather metrics
			metrics, err := a.GetMetricer.GetMetrics(deployment)
			if err != nil {
				log.Printf("Failed to gather metrics\n%v", err)
				break
			}

			// Evaluate based on metrics
			evaluation, err := a.GetEvaluationer.GetEvaluation(metrics)
			if err != nil {
				log.Printf("Failed to evaluate metrics\n%v", err)
				break
			}

			// Check if evaluation requires an action
			// If the deployment needs scaled up/down
			if evaluation.TargetReplicas != deployment.Spec.Replicas {
				// Scale deployment
				deployment.Spec.Replicas = evaluation.TargetReplicas
				_, err = a.DeploymentsClient.Update(deployment)
				if err != nil {
					log.Fatalf(err.Error())
				}
			}
		}
	}
}
