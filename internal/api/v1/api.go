/*
Copyright 2021 The Custom Pod Autoscaler Authors.

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

// Package v1 provides routing and endpoints for the Custom Pod Autoscaler
// HTTP REST API version 1. Endpoints implemented as handlers, errors returned as
// valid JSON.
package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	apiv1 "github.com/jthomperoo/custom-pod-autoscaler/api/v1"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/evaluatecalc"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/metricget"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/resourceclient"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/scaling"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/scale"
)

const (
	dryRunQueryParam = "dry_run"
)

// API is the Custom Pod Autoscaler REST API, exposing endpoints to retrieve metrics/evaluations
type API struct {
	Router          chi.Router
	Config          *config.Config
	Client          resourceclient.Client
	Scaler          scaling.Scaler
	GetMetricer     metricget.GetMetricer
	GetEvaluationer evaluatecalc.GetEvaluationer
}

// Routes sets up routing for the API
func (api *API) Routes() {
	// Set up routing
	api.Router.Route("/api/v1", func(r chi.Router) {
		r.NotFound(api.notFound)
		r.MethodNotAllowed(api.methodNotAllowed)
		r.Get("/metrics", api.getMetrics)
		r.Post("/evaluation", api.getEvaluation)
	})
}

func (api *API) getMetrics(w http.ResponseWriter, r *http.Request) {
	// Determine if it is a dry run
	dryRun := true
	dryRunParam := r.URL.Query().Get(dryRunQueryParam)
	if dryRunParam == "" {
		dryRun = false
	} else {
		b, err := strconv.ParseBool(dryRunParam)
		if err != nil {
			apiError(w, &apiv1.Error{
				Message: fmt.Sprintf("Invalid format for 'dry_run' query parameter; '%s' is not a valid boolean value", dryRunParam),
				Code:    http.StatusBadRequest,
			})
			return
		}
		dryRun = b
	}

	// Get resource being managed
	resource, err := api.Client.Get(api.Config.ScaleTargetRef.APIVersion, api.Config.ScaleTargetRef.Kind, api.Config.ScaleTargetRef.Name, api.Config.Namespace)
	if err != nil {
		apiError(w, &apiv1.Error{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Set run type
	runType := config.APIRunType
	if dryRun {
		runType = config.APIDryRunRunType
	}

	// Get metrics
	metrics, err := api.GetMetricer.GetMetrics(metric.Info{
		Resource: resource,
		RunType:  runType,
	})
	if err != nil {
		apiError(w, &apiv1.Error{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Convert metrics into JSON
	response, err := json.Marshal(metrics)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (api *API) getEvaluation(w http.ResponseWriter, r *http.Request) {
	// Determine if it is a dry run
	dryRun := true
	dryRunParam := r.URL.Query().Get(dryRunQueryParam)
	if dryRunParam == "" {
		dryRun = false
	} else {
		b, err := strconv.ParseBool(dryRunParam)
		if err != nil {
			apiError(w, &apiv1.Error{
				Message: fmt.Sprintf("Invalid format for 'dry_run' query parameter; '%s' is not a valid boolean value", dryRunParam),
				Code:    http.StatusBadRequest,
			})
			return
		}
		dryRun = b
	}

	// Get resource being managed
	resource, err := api.Client.Get(api.Config.ScaleTargetRef.APIVersion, api.Config.ScaleTargetRef.Kind, api.Config.ScaleTargetRef.Name, api.Config.Namespace)
	if err != nil {
		apiError(w, &apiv1.Error{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Set run type
	runType := config.APIRunType
	if dryRun {
		runType = config.APIDryRunRunType
	}

	// Get metrics
	metrics, err := api.GetMetricer.GetMetrics(metric.Info{
		Resource: resource,
		RunType:  runType,
	})
	if err != nil {
		apiError(w, &apiv1.Error{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Get evaluations for metrics
	evaluation, err := api.GetEvaluationer.GetEvaluation(evaluate.Info{
		Metrics:  metrics,
		Resource: resource,
		RunType:  runType,
	})
	if err != nil {
		apiError(w, &apiv1.Error{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Scale if not a dry run
	if !dryRun {
		evaluation, err = api.Scaler.Scale(scale.Info{
			Evaluation:     *evaluation,
			Resource:       resource,
			MinReplicas:    api.Config.MinReplicas,
			MaxReplicas:    api.Config.MaxReplicas,
			Namespace:      api.Config.Namespace,
			ScaleTargetRef: api.Config.ScaleTargetRef,
			RunType:        runType,
		})
		if err != nil {
			apiError(w, &apiv1.Error{
				Message: err.Error(),
				Code:    http.StatusInternalServerError,
			})
			return
		}
	}

	// Convert evaluation into JSON
	response, err := json.Marshal(evaluation)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (api *API) notFound(w http.ResponseWriter, r *http.Request) {
	apiError(w, &apiv1.Error{
		Message: fmt.Sprintf("Resource '%s' not found", r.URL.Path),
		Code:    http.StatusNotFound,
	})
}

func (api *API) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	apiError(w, &apiv1.Error{
		Message: fmt.Sprintf("Method '%s' not allowed on resource '%s'", r.Method, r.URL.Path),
		Code:    http.StatusMethodNotAllowed,
	})
}

func apiError(w http.ResponseWriter, apiErr *apiv1.Error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Convert into JSON
	response, err := json.Marshal(apiErr)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}
	w.WriteHeader(apiErr.Code)
	w.Write(response)
}
