/*
Copyright 2020 The Custom Pod Autoscaler Authors.

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
// +build unit

package v1_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"

	"github.com/google/go-cmp/cmp"

	apiv1 "github.com/jthomperoo/custom-pod-autoscaler/api/v1"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/resourceclient"
	"github.com/jthomperoo/custom-pod-autoscaler/scale"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type failGetMetrics struct{}

func (f *failGetMetrics) GetMetrics(spec metric.Spec) (*metric.ResourceMetrics, error) {
	return nil, errors.New("FAIL GET METRICS")
}

type successGetMetrics struct{}

func (s *successGetMetrics) GetMetrics(spec metric.Spec) (*metric.ResourceMetrics, error) {
	return &metric.ResourceMetrics{
		Metrics: []*metric.Metric{
			&metric.Metric{
				Value:    "SUCCESS",
				Resource: "SUCCESS_POD",
			},
		},
		Resource: spec.Resource,
	}, nil
}

type failGetEvaluation struct{}

func (f *failGetEvaluation) GetEvaluation(spec evaluate.Spec) (*evaluate.Evaluation, error) {
	return nil, errors.New("FAIL GET EVALUATION")
}

type successGetEvaluation struct{}

func (s *successGetEvaluation) GetEvaluation(spec evaluate.Spec) (*evaluate.Evaluation, error) {
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
		getMetricer      metric.GetMetricer
		getEvaluationer  evaluate.GetEvaluationer
		scaler           scale.Scaler
	}{
		{
			"Fail to get resource metric gathering",
			`{"message":"fail getting resource","code":500}`,
			http.StatusInternalServerError,
			"GET",
			"/api/v1/metrics",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
			strings.ReplaceAll(strings.ReplaceAll(`
			{
				"metrics":[
					{
						"resource":"SUCCESS_POD",
						"value":"SUCCESS"
					}
				],
				"resource":{
					"metadata":{
						"name":"test",
						"creationTimestamp":null
					},
					"spec":{
						"selector":null,
						"template":{
							"metadata":{
								"creationTimestamp":null
							},
							"spec":{
								"containers":null
							}
						},
						"strategy":{}
					},
					"status":{}
				}
			}`, "\n", ""), "\t", ""),
			http.StatusOK,
			"GET",
			"/api/v1/metrics",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
			strings.ReplaceAll(strings.ReplaceAll(`
			{
				"metrics":[
					{
						"resource":"SUCCESS_POD",
						"value":"SUCCESS"
					}
				],
				"resource":{
					"metadata":{
						"name":"test",
						"creationTimestamp":null
					},
					"spec":{
						"selector":null,
						"template":{
							"metadata":{
								"creationTimestamp":null
							},
							"spec":{
								"containers":null
							}
						},
						"strategy":{}
					},
					"status":{}
				}
			}`, "\n", ""), "\t", ""),
			http.StatusOK,
			"GET",
			"/api/v1/metrics?dry_run=false",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
			strings.ReplaceAll(strings.ReplaceAll(`
			{
				"metrics":[
					{
						"resource":"SUCCESS_POD",
						"value":"SUCCESS"
					}
				],
				"resource":{
					"metadata":{
						"name":"test",
						"creationTimestamp":null
					},
					"spec":{
						"selector":null,
						"template":{
							"metadata":{
								"creationTimestamp":null
							},
							"spec":{
								"containers":null
							}
						},
						"strategy":{}
					},
					"status":{}
				}
			}`, "\n", ""), "\t", ""),
			http.StatusOK,
			"GET",
			"/api/v1/metrics?dry_run=true",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
				ScaleReactor: func(spec scale.Spec) (*evaluate.Evaluation, error) {
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
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
				ScaleReactor: func(spec scale.Spec) (*evaluate.Evaluation, error) {
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
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
				ScaleReactor: func(spec scale.Spec) (*evaluate.Evaluation, error) {
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
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
				ScaleTargetRef: &v1.CrossVersionObjectReference{
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
