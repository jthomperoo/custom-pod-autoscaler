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
	"github.com/jthomperoo/custom-pod-autoscaler/cpatest"
	"github.com/jthomperoo/custom-pod-autoscaler/models"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/autoscaling/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type getMetricer interface {
	GetMetrics(deployment *appsv1.Deployment) (*models.ResourceMetrics, error)
}

type getEvaluationer interface {
	GetEvaluation(resourceMetrics *models.ResourceMetrics) (*models.Evaluation, error)
}

func TestAPI(t *testing.T) {
	var tests = []struct {
		description      string
		expectedResponse string
		expectedCode     int
		method           string
		endpoint         string
		config           *config.Config
		clientset        kubernetes.Interface
		getMetricer      getMetricer
		getEvaluationer  getEvaluationer
	}{
		{
			"Get metrics no deployments",
			"{\"message\":\"deployments.apps \\\"test\\\" not found\",\"code\":500}",
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
			fake.NewSimpleClientset(),
			nil,
			nil,
		},
		{
			"Get metrics fail metric gathering",
			"{\"message\":\"FAIL GET METRICS\",\"code\":500}",
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
			fake.NewSimpleClientset(
				cpatest.Deployment("test", "test-namespace", nil),
			),
			&failGetMetrics{},
			nil,
		},
		{
			"Get metrics success metric gathering",
			"{\"deployment\":\"test\",\"metrics\":[{\"pod\":\"SUCCESS_POD\",\"value\":\"SUCCESS\"}]}",
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
			fake.NewSimpleClientset(
				cpatest.Deployment("test", "test-namespace", nil),
			),
			&successGetMetrics{},
			nil,
		},
		{
			"Get metrics success metric gathering two deployment same namespace",
			"{\"deployment\":\"target\",\"metrics\":[{\"pod\":\"SUCCESS_POD\",\"value\":\"SUCCESS\"}]}",
			http.StatusOK,
			"GET",
			"/metrics",
			&config.Config{
				Namespace: "test-namespace",
				ScaleTargetRef: &v1.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "target",
					APIVersion: "apps/v1",
				},
			},
			fake.NewSimpleClientset(
				cpatest.Deployment("target", "test-namespace", nil),
				cpatest.Deployment("not target", "test-namespace", nil),
			),
			&successGetMetrics{},
			nil,
		},
		{
			"Get metrics success metric gathering two deployment different namespaces",
			"{\"deployment\":\"target\",\"metrics\":[{\"pod\":\"SUCCESS_POD\",\"value\":\"SUCCESS\"}]}",
			http.StatusOK,
			"GET",
			"/metrics",
			&config.Config{
				Namespace: "target-namespace",
				ScaleTargetRef: &v1.CrossVersionObjectReference{
					Kind:       "deployment",
					Name:       "target",
					APIVersion: "apps/v1",
				},
			},
			fake.NewSimpleClientset(
				cpatest.Deployment("target", "target-namespace", nil),
				cpatest.Deployment("target", "not-target-namespace", nil),
			),
			&successGetMetrics{},
			nil,
		},
		{
			"Get evaluation no deployments",
			"{\"message\":\"deployments.apps \\\"test\\\" not found\",\"code\":500}",
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
			fake.NewSimpleClientset(),
			nil,
			nil,
		},
		{
			"Get evaluation fail evaluate",
			"{\"message\":\"FAIL GET EVALUATION\",\"code\":500}",
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
			fake.NewSimpleClientset(
				cpatest.Deployment("test", "test-namespace", nil),
			),
			&successGetMetrics{},
			&failGetEvaluation{},
		},
		{
			"Get evaluation fail metric gathering",
			"{\"message\":\"FAIL GET METRICS\",\"code\":500}",
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
			fake.NewSimpleClientset(
				cpatest.Deployment("test", "test-namespace", nil),
			),
			&failGetMetrics{},
			&successGetEvaluation{},
		},
		{
			"Get evaluation success evaluate",
			"{\"target_replicas\":1}",
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
			fake.NewSimpleClientset(
				cpatest.Deployment("test", "test-namespace", nil),
			),
			&successGetMetrics{},
			&successGetEvaluation{},
		},
		{
			"Non existent endpoint",
			"{\"message\":\"Resource '/non_existant' not found\",\"code\":404}",
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
			fake.NewSimpleClientset(),
			nil,
			nil,
		},
		{
			"Use incorrect method on get metrics endpoint",
			"{\"message\":\"Method 'DELETE' not allowed on resource '/metrics'\",\"code\":405}",
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
			fake.NewSimpleClientset(),
			nil,
			nil,
		},
		{
			"Use incorrect method on evaluation endpoint",
			"{\"message\":\"Method 'POST' not allowed on resource '/evaluation'\",\"code\":405}",
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
			fake.NewSimpleClientset(),
			nil,
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			api := &api.API{
				Router:            chi.NewRouter(),
				Config:            test.config,
				Clientset:         test.clientset,
				DeploymentsClient: test.clientset.AppsV1().Deployments(test.config.Namespace),
				GetMetricer:       test.getMetricer,
				GetEvaluationer:   test.getEvaluationer,
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

type failGetMetrics struct{}

func (f *failGetMetrics) GetMetrics(deployment *appsv1.Deployment) (*models.ResourceMetrics, error) {
	return nil, errors.New("FAIL GET METRICS")
}

type successGetMetrics struct{}

func (s *successGetMetrics) GetMetrics(deployment *appsv1.Deployment) (*models.ResourceMetrics, error) {
	return &models.ResourceMetrics{
		DeploymentName: deployment.Name,
		Metrics: []*models.Metric{
			&models.Metric{
				Value: "SUCCESS",
				Pod:   "SUCCESS_POD",
			},
		},
		Deployment: deployment,
	}, nil
}

type failGetEvaluation struct{}

func (f *failGetEvaluation) GetEvaluation(resourceMetrics *models.ResourceMetrics) (*models.Evaluation, error) {
	return nil, errors.New("FAIL GET EVALUATION")
}

type successGetEvaluation struct{}

func (s *successGetEvaluation) GetEvaluation(resourceMetrics *models.ResourceMetrics) (*models.Evaluation, error) {
	targetReplicas := int32(1)
	return &models.Evaluation{
		TargetReplicas: &targetReplicas,
	}, nil
}
