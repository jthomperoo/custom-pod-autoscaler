//go:build unit
// +build unit

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

package v1_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"

	"github.com/google/go-cmp/cmp"

	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	apiv1 "github.com/jthomperoo/custom-pod-autoscaler/v2/internal/api/v1"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/evaluatecalc"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/metricget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/resourceclient"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/scaling"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/scale"

	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type failGetMetrics struct{}

func (f *failGetMetrics) GetMetrics(info metric.Info) ([]*metric.ResourceMetric, error) {
	return nil, errors.New("FAIL GET METRICS")
}

type successGetMetrics struct{}

func (s *successGetMetrics) GetMetrics(info metric.Info) ([]*metric.ResourceMetric, error) {
	return []*metric.ResourceMetric{
		{
			Value:    "SUCCESS",
			Resource: "SUCCESS_POD",
		},
	}, nil
}

type failGetEvaluation struct{}

func (f *failGetEvaluation) GetEvaluation(info evaluate.Info) (*evaluate.Evaluation, error) {
	return nil, errors.New("FAIL GET EVALUATION")
}

type successGetEvaluation struct{}

func (s *successGetEvaluation) GetEvaluation(info evaluate.Info) (*evaluate.Evaluation, error) {
	return &evaluate.Evaluation{
		TargetReplicas: int32(1),
	}, nil
}

func TestAPI(t *testing.T) {
	var tests = []struct {
		description      string
		expectedResponse string
		expectedCode     int
		method           string
		endpoint         string
		config           *config.Config
		client           resourceclient.Client
		getMetricer      metricget.GetMetricer
		getEvaluationer  evaluatecalc.GetEvaluationer
		scaler           scaling.Scaler
	}{
		{
			"Fail to get resource metric gathering",
			`{"message":"fail getting resource","code":500}`,
			http.StatusInternalServerError,
			"GET",
			"/api/v1/metrics",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return nil, errors.New("fail getting resource")
				},
			},
			nil,
			nil,
			nil,
		},
		{
			"Get metrics fail metric gathering",
			`{"message":"FAIL GET METRICS","code":500}`,
			http.StatusInternalServerError,
			"GET",
			"/api/v1/metrics",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: name,
						},
					}, nil
				},
			},
			&failGetMetrics{},
			nil,
			nil,
		},
		{
			"Get metrics fail invalid dry_run parameter",
			`{"message":"Invalid format for 'dry_run' query parameter; 'invalid' is not a valid boolean value","code":400}`,
			http.StatusBadRequest,
			"GET",
			"/api/v1/metrics?dry_run=invalid",
			nil,
			nil,
			nil,
			nil,
			nil,
		},
		{
			"Get metrics success metric gathering, not dry run, no parameter provided",
			`[{"resource":"SUCCESS_POD","value":"SUCCESS"}]`,
			http.StatusOK,
			"GET",
			"/api/v1/metrics",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: name,
						},
					}, nil
				},
			},
			&successGetMetrics{},
			nil,
			nil,
		},
		{
			"Get metrics success metric gathering, not dry run, parameter provided",
			`[{"resource":"SUCCESS_POD","value":"SUCCESS"}]`,
			http.StatusOK,
			"GET",
			"/api/v1/metrics?dry_run=false",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: name,
						},
					}, nil
				},
			},
			&successGetMetrics{},
			nil,
			nil,
		},
		{
			"Get metrics success metric gathering, dry run",
			`[{"resource":"SUCCESS_POD","value":"SUCCESS"}]`,
			http.StatusOK,
			"GET",
			"/api/v1/metrics?dry_run=true",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: name,
						},
					}, nil
				},
			},
			&successGetMetrics{},
			nil,
			nil,
		},
		{
			"Evaluate fail invalid dry_run parameter",
			`{"message":"Invalid format for 'dry_run' query parameter; 'invalid' is not a valid boolean value","code":400}`,
			http.StatusBadRequest,
			"POST",
			"/api/v1/evaluation?dry_run=invalid",
			nil,
			nil,
			nil,
			nil,
			nil,
		},
		{
			"Evaluate fail to get resource",
			`{"message":"fail to get resource","code":500}`,
			http.StatusInternalServerError,
			"POST",
			"/api/v1/evaluation",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return nil, errors.New("fail to get resource")
				},
			},
			nil,
			nil,
			nil,
		},
		{
			"Evaluate fail to get metrics",
			`{"message":"FAIL GET METRICS","code":500}`,
			http.StatusInternalServerError,
			"POST",
			"/api/v1/evaluation",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: name,
						},
					}, nil
				},
			},
			&failGetMetrics{},
			&successGetEvaluation{},
			nil,
		},
		{
			"Evaluate fail to get evaluation",
			`{"message":"FAIL GET EVALUATION","code":500}`,
			http.StatusInternalServerError,
			"POST",
			"/api/v1/evaluation",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: name,
						},
					}, nil
				},
			},
			&successGetMetrics{},
			&failGetEvaluation{},
			nil,
		},
		{
			"Evaluate fail failure scaling",
			`{"message":"FAILURE SCALING","code":500}`,
			http.StatusInternalServerError,
			"POST",
			"/api/v1/evaluation",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						TypeMeta: metav1.TypeMeta{
							Kind:       "deployment",
							APIVersion: "apps/v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "test",
						},
					}, nil
				},
			},
			&successGetMetrics{},
			&successGetEvaluation{},
			&fake.Scaler{
				ScaleReactor: func(info scale.Info) (*evaluate.Evaluation, error) {
					return nil, errors.New("FAILURE SCALING")
				},
			},
		},
		{
			"Evaluate success, not dry run, no parameter provided",
			`{"targetReplicas":1}`,
			http.StatusOK,
			"POST",
			"/api/v1/evaluation",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: name,
						},
					}, nil
				},
			},
			&successGetMetrics{},
			&successGetEvaluation{},
			&fake.Scaler{
				ScaleReactor: func(info scale.Info) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: 1,
					}, nil
				},
			},
		},
		{
			"Evaluate success, not dry run, parameter provided",
			`{"targetReplicas":1}`,
			http.StatusOK,
			"POST",
			"/api/v1/evaluation?dry_run=false",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: name,
						},
					}, nil
				},
			},
			&successGetMetrics{},
			&successGetEvaluation{},
			&fake.Scaler{
				ScaleReactor: func(info scale.Info) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: 1,
					}, nil
				},
			},
		},
		{
			"Evaluate success, dry run",
			`{"targetReplicas":1}`,
			http.StatusOK,
			"POST",
			"/api/v1/evaluation?dry_run=true",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{
				GetReactor: func(apiVersion, kind, name, namespace string) (metav1.Object, error) {
					return &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: name,
						},
					}, nil
				},
			},
			&successGetMetrics{},
			&successGetEvaluation{},
			nil,
		},
		{
			"Non existent endpoint",
			`{"message":"Resource '/api/v1/non_existent' not found","code":404}`,
			http.StatusNotFound,
			"GET",
			"/api/v1/non_existent",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			&fake.ResourceClient{},
			nil,
			nil,
			nil,
		},
		{
			"Use incorrect method on metrics endpoint",
			`{"message":"Method 'DELETE' not allowed on resource '/api/v1/metrics'","code":405}`,
			http.StatusMethodNotAllowed,
			"DELETE",
			"/api/v1/metrics",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			nil,
			nil,
			nil,
			nil,
		},
		{
			"Use incorrect method on evaluation endpoint",
			`{"message":"Method 'DELETE' not allowed on resource '/api/v1/evaluation'","code":405}`,
			http.StatusMethodNotAllowed,
			"DELETE",
			"/api/v1/evaluation",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &autoscaling.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "test",
					APIVersion: "apps/v1",
				},
			},
			nil,
			nil,
			nil,
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			api := &apiv1.API{
				Router:          chi.NewRouter(),
				Config:          test.config,
				Client:          test.client,
				GetMetricer:     test.getMetricer,
				GetEvaluationer: test.getEvaluationer,
				Scaler:          test.scaler,
			}
			api.Routes()
			req, err := http.NewRequest(test.method, test.endpoint, nil)
			if err != nil {
				t.Error(err)
			}
			recorder := httptest.NewRecorder()
			api.Router.ServeHTTP(recorder, req)

			if !cmp.Equal(recorder.Code, test.expectedCode) {
				t.Errorf("response code mismatch (-want +got):\n%s", cmp.Diff(test.expectedCode, recorder.Code))
				return
			}

			if !cmp.Equal(recorder.Body.String(), test.expectedResponse) {
				t.Errorf("response code mismatch (-want +got):\n%s", cmp.Diff(test.expectedResponse, recorder.Body.String()))
				return
			}
		})
	}
}
