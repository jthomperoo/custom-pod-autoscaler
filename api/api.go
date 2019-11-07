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

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/models"
	appsv1 "k8s.io/api/apps/v1"
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

// API is the Custom Pod Autoscaler REST API, exposing endpoints to retrieve metrics/evaluations
type API struct {
	Config            *config.Config
	Clientset         *kubernetes.Clientset
	DeploymentsClient v1.DeploymentInterface
	GetMetricer       getMetricer
	GetEvaluationer   getEvaluationer
}

// Start sets up routing for the API and starts listening
func (api *API) Start() {
	// Set up routing
	r := chi.NewRouter()
	r.Get("/metrics", api.getMetrics)
	r.Get("/evaluation", api.getEvaluation)

	// Set up server
	srv := http.Server{Addr: fmt.Sprintf("%s:%d", api.Config.Host, api.Config.Port), Handler: r}

	// Set up channel for handling shutdown requests
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Handle shutdowns
	go func() {
		for range shutdown {
			log.Println("Shutting down...")
			// Immediate shutdown
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			srv.Shutdown(ctx)
		}
	}()

	// Start API
	srv.ListenAndServe()
}

func (api *API) getMetrics(w http.ResponseWriter, r *http.Request) {
	// Get deployments being managed
	deployment, err := api.Clientset.AppsV1().Deployments(api.Config.Namespace).Get(api.Config.ScaleTargetRef.Name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	// Get metrics
	metrics, err := api.GetMetricer.GetMetrics(deployment)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Convert metrics into JSON
	response, err := json.Marshal(metrics)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (api *API) getEvaluation(w http.ResponseWriter, r *http.Request) {
	// Get deployments being managed
	deployment, err := api.Clientset.AppsV1().Deployments(api.Config.Namespace).Get(api.Config.ScaleTargetRef.Name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	// Get metrics
	metrics, err := api.GetMetricer.GetMetrics(deployment)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Get evaluations for metrics
	evaluations, err := api.GetEvaluationer.GetEvaluation(metrics)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Convert evaluations into JSON
	response, err := json.Marshal(evaluations)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
