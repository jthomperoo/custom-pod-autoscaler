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

// Package api provides routing and endpoints for the Custom Pod Autoscaler
// HTTP REST API. Endpoints implemented as handlers, errors returned as valid
// JSON.
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RunType api marks the metric gathering/evaluation as running during an API request
const RunType = "api"

type getMetricer interface {
	GetMetrics(deployment *appsv1.Deployment) (*metric.ResourceMetrics, error)
}

type getEvaluationer interface {
	GetEvaluation(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error)
}

// Error is an error response from the API, with the status code and an error message
type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// API is the Custom Pod Autoscaler REST API, exposing endpoints to retrieve metrics/evaluations
type API struct {
	Router          chi.Router
	Config          *config.Config
	Clientset       kubernetes.Interface
	GetMetricer     getMetricer
	GetEvaluationer getEvaluationer
}

// Routes sets up routing for the API
func (api *API) Routes() {
	// Set up routing
	api.Router.Get("/metrics", api.getMetrics)
	api.Router.Get("/evaluation", api.getEvaluation)
	api.Router.NotFound(api.notFound)
	api.Router.MethodNotAllowed(api.methodNotAllowed)
}

func (api *API) getMetrics(w http.ResponseWriter, r *http.Request) {
	// Get deployments being managed
	deployment, err := api.Clientset.AppsV1().Deployments(api.Config.Namespace).Get(api.Config.ScaleTargetRef.Name, metav1.GetOptions{})
	if err != nil {
		apiError(w, &Error{
			err.Error(),
			http.StatusInternalServerError,
		})
		return
	}
	// Get metrics
	metrics, err := api.GetMetricer.GetMetrics(deployment)
	if err != nil {
		apiError(w, &Error{
			err.Error(),
			http.StatusInternalServerError,
		})
		return
	}
	metrics.RunType = RunType
	// Convert metrics into JSON
	response, err := json.Marshal(metrics)
	if err != nil {
		// Should not occur, panic
		log.Panic(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (api *API) getEvaluation(w http.ResponseWriter, r *http.Request) {
	// Get deployments being managed
	deployment, err := api.Clientset.AppsV1().Deployments(api.Config.Namespace).Get(api.Config.ScaleTargetRef.Name, metav1.GetOptions{})
	if err != nil {
		apiError(w, &Error{
			err.Error(),
			http.StatusInternalServerError,
		})
		return
	}
	// Get metrics
	metrics, err := api.GetMetricer.GetMetrics(deployment)
	if err != nil {
		apiError(w, &Error{
			err.Error(),
			http.StatusInternalServerError,
		})
		return
	}
	metrics.RunType = RunType
	// Get evaluations for metrics
	evaluations, err := api.GetEvaluationer.GetEvaluation(metrics)
	if err != nil {
		apiError(w, &Error{
			err.Error(),
			http.StatusInternalServerError,
		})
		return
	}
	// Convert evaluations into JSON
	response, err := json.Marshal(evaluations)
	if err != nil {
		// Should not occur, panic
		log.Panic(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (api *API) notFound(w http.ResponseWriter, r *http.Request) {
	apiError(w, &Error{
		Message: fmt.Sprintf("Resource '%s' not found", r.URL.Path),
		Code:    http.StatusNotFound,
	})
}

func (api *API) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	apiError(w, &Error{
		Message: fmt.Sprintf("Method '%s' not allowed on resource '%s'", r.Method, r.URL.Path),
		Code:    http.StatusMethodNotAllowed,
	})
}

func apiError(w http.ResponseWriter, apiErr *Error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Convert into JSON
	response, err := json.Marshal(apiErr)
	if err != nil {
		// Should not occur, panic
		log.Panic(err)
	}
	w.WriteHeader(apiErr.Code)
	w.Write(response)
}
