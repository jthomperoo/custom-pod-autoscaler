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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/scaler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

// ConfigureAPI sets up endpoints and begins listening for API requests
func ConfigureAPI(clientset *kubernetes.Clientset, deploymentsClient v1.DeploymentInterface, config *config.Config) {
	// Set up shared resources
	api := &api{
		clientset:         clientset,
		deploymentsClient: deploymentsClient,
		config:            config,
	}

	// Set up routing and serve API
	r := chi.NewRouter()
	r.Get("/metrics", api.getMetrics)
	r.Get("/evaluations", api.getEvaluations)
	http.ListenAndServe(fmt.Sprintf("%s:%d", config.Host, config.Port), r)
}

type api struct {
	config            *config.Config
	clientset         *kubernetes.Clientset
	deploymentsClient v1.DeploymentInterface
}

func (api *api) getMetrics(w http.ResponseWriter, r *http.Request) {
	// Get deployments being managed
	deployments, err := api.deploymentsClient.List(metav1.ListOptions{LabelSelector: api.config.Selector})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	// Get metrics
	metrics, err := scaler.GetMetrics(api.clientset, deployments, api.config)
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

func (api *api) getEvaluations(w http.ResponseWriter, r *http.Request) {
	// Get deployments being managed
	deployments, err := api.deploymentsClient.List(metav1.ListOptions{LabelSelector: api.config.Selector})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	// Get metrics
	metrics, err := scaler.GetMetrics(api.clientset, deployments, api.config)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Get evaluations for metrics
	evaluations, err := scaler.GetEvaluations(metrics, api.config)
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
