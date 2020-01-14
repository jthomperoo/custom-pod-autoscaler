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
// +build unit

package api_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"

	"github.com/google/go-cmp/cmp"

	"github.com/jthomperoo/custom-pod-autoscaler/api"
	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/custom-pod-autoscaler/resourceclient"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type failGetMetrics struct{}

func (f *failGetMetrics) GetMetrics(resource metav1.Object) (*metric.ResourceMetrics, error) {
	return nil, errors.New("FAIL GET METRICS")
}

type successGetMetrics struct{}

func (s *successGetMetrics) GetMetrics(resource metav1.Object) (*metric.ResourceMetrics, error) {
	return &metric.ResourceMetrics{
		ResourceName: resource.GetName(),
		Metrics: []*metric.Metric{
			&metric.Metric{
				Value:    "SUCCESS",
				Resource: "SUCCESS_POD",
			},
		},
		Resource: resource,
	}, nil
}

type failGetEvaluation struct{}

func (f *failGetEvaluation) GetEvaluation(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
	return nil, errors.New("FAIL GET EVALUATION")
}

type successGetEvaluation struct{}

func (s *successGetEvaluation) GetEvaluation(resourceMetrics *metric.ResourceMetrics) (*evaluate.Evaluation, error) {
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
	}{
		{
			"Fail to get resource",
			`{"message":"fail getting resource","code":500}`,
			http.StatusInternalServerError,
			"GET",
			"/metrics",
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
		},
		{
			"Get metrics fail metric gathering",
			`{"message":"FAIL GET METRICS","code":500}`,
			http.StatusInternalServerError,
			"GET",
			"/metrics",
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
		},
		{
			"Get metrics success metric gathering",
			`{"resource":"test","run_type":"api","metrics":[{"resource":"SUCCESS_POD","value":"SUCCESS"}]}`,
			http.StatusOK,
			"GET",
			"/metrics",
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
		},
		{
			"Fail to get resource",
			`{"message":"fail to get resource","code":500}`,
			http.StatusInternalServerError,
			"GET",
			"/evaluation",
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
		},
		{
			"Get evaluation fail evaluate",
			`{"message":"FAIL GET EVALUATION","code":500}`,
			http.StatusInternalServerError,
			"GET",
			"/evaluation",
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
		},
		{
			"Get evaluation fail metric gathering",
			`{"message":"FAIL GET METRICS","code":500}`,
			http.StatusInternalServerError,
			"GET",
			"/evaluation",
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
		},
		{
			"Get evaluation success evaluate",
			`{"target_replicas":1}`,
			http.StatusOK,
			"GET",
			"/evaluation",
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
		},
		{
			"Non existent endpoint",
			`{"message":"Resource '/non_existant' not found","code":404}`,
			http.StatusNotFound,
			"GET",
			"/non_existant",
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
		},
		{
			"Use incorrect method on get metrics endpoint",
			`{"message":"Method 'DELETE' not allowed on resource '/metrics'","code":405}`,
			http.StatusMethodNotAllowed,
			"DELETE",
			"/metrics",
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
		},
		{
			"Use incorrect method on evaluation endpoint",
			`{"message":"Method 'POST' not allowed on resource '/evaluation'","code":405}`,
			http.StatusMethodNotAllowed,
			"POST",
			"/evaluation",
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
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			api := &api.API{
				Router:          chi.NewRouter(),
				Config:          test.config,
				Client:          test.client,
				GetMetricer:     test.getMetricer,
				GetEvaluationer: test.getEvaluationer,
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
