/*
Copyright 2022 The Custom Pod Autoscaler Authors.

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

package k8smetricget_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/externalget"
	metricsclient "github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/metrics"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/objectget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/podsget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/podutil"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/k8smetricget/resourceget"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/external"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/object"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/podmetrics"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/pods"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/resource"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/k8smetric/value"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	corelisters "k8s.io/client-go/listers/core/v1"
)

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestGetMetrics(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description   string
		expected      []*k8smetric.Metric
		expectedErr   error
		resource      resourceget.Gatherer
		object        objectget.Gatherer
		pods          podsget.Gatherer
		external      externalget.Gatherer
		deployment    metav1.Object
		specs         []config.K8sMetricSpec
		namespace     string
		scaleResource *autoscalingv1.Scale
	}{
		{
			"Single unknown metric type",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: unknown metric source type "invalid"`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
			[]config.K8sMetricSpec{
				{
					Type: "invalid",
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single object metric, fail to convert label",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: failed to get object metric: "invalid" is not a valid pod selector operator`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &config.K8sObjectMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Selector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Operator: "invalid",
									},
								},
							},
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single object metric, invalid target",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: invalid object metric source: must be either value or average value`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &config.K8sObjectMetricSource{
						Target: config.K8sMetricTarget{
							Type: "invalid",
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single object metric, value metric, fail to get metric",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: failed to get object metric: fail to get object metric`),
			nil,
			&fake.ObjectGatherer{
				GetMetricReactor: func(metricName, namespace string, objectRef *autoscaling.CrossVersionObjectReference, selector, metricSelector labels.Selector) (*object.Metric, error) {
					return nil, errors.New("fail to get object metric")
				},
			},
			nil,
			nil,
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &config.K8sObjectMetricSource{
						Target: config.K8sMetricTarget{
							Type: autoscaling.ValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single object metric, deployment, value metric, success",
			[]*k8smetric.Metric{
				{
					CurrentReplicas: 1,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.ObjectMetricSourceType,
						Object: &config.K8sObjectMetricSource{
							Target: config.K8sMetricTarget{
								Type: autoscaling.ValueMetricType,
							},
						},
					},
					Object: &object.Metric{
						Current: value.MetricValue{
							Value: int64Ptr(5),
						},
						ReadyPodCount: int64Ptr(2),
					},
				},
			},
			nil,
			nil,
			&fake.ObjectGatherer{
				GetMetricReactor: func(metricName, namespace string, objectRef *autoscaling.CrossVersionObjectReference, selector, metricSelector labels.Selector) (*object.Metric, error) {
					return &object.Metric{
						Current: value.MetricValue{
							Value: int64Ptr(5),
						},
						ReadyPodCount: int64Ptr(2),
					}, nil
				},
			},
			nil,
			nil,
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"test": "test",
						},
					},
				},
			},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &config.K8sObjectMetricSource{
						Target: config.K8sMetricTarget{
							Type: autoscaling.ValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 1,
				},
			},
		},
		{
			"Single object metric, argo rollout, value metric, success",
			[]*k8smetric.Metric{
				{
					CurrentReplicas: 1,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.ObjectMetricSourceType,
						Object: &config.K8sObjectMetricSource{
							Target: config.K8sMetricTarget{
								Type: autoscaling.ValueMetricType,
							},
						},
					},
					Object: &object.Metric{
						Current: value.MetricValue{
							Value: int64Ptr(5),
						},
						ReadyPodCount: int64Ptr(2),
					},
				},
			},
			nil,
			nil,
			&fake.ObjectGatherer{
				GetMetricReactor: func(metricName, namespace string, objectRef *autoscaling.CrossVersionObjectReference, selector, metricSelector labels.Selector) (*object.Metric, error) {
					return &object.Metric{
						Current: value.MetricValue{
							Value: int64Ptr(5),
						},
						ReadyPodCount: int64Ptr(2),
					}, nil
				},
			},
			nil,
			nil,
			&unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"selector": map[string]interface{}{
							"matchLabels": map[string]string{
								"test": "test",
							},
						},
						"replicas": 1,
					},
				},
			},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &config.K8sObjectMetricSource{
						Target: config.K8sMetricTarget{
							Type: autoscaling.ValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 1,
				},
			},
		},
		{
			"Single object metric, replicaset, average value metric, fail to get metric",
			nil,
			errors.New("invalid metrics (1 invalid out of 1), first error is: failed to get object metric: fail to get object metric"),
			nil,
			&fake.ObjectGatherer{
				GetPerPodMetricReactor: func(metricName, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (*object.Metric, error) {
					return nil, errors.New("fail to get object metric")
				},
			},
			nil,
			nil,
			&appsv1.ReplicaSet{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &config.K8sObjectMetricSource{
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single object metric, statefulset, average value metric, success",
			[]*k8smetric.Metric{
				{
					CurrentReplicas: 3,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.ObjectMetricSourceType,
						Object: &config.K8sObjectMetricSource{
							Target: config.K8sMetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Object: &object.Metric{
						Current: value.MetricValue{
							AverageValue: int64Ptr(17),
						},
						ReadyPodCount: int64Ptr(5),
					},
				},
			},
			nil,
			nil,
			&fake.ObjectGatherer{
				GetPerPodMetricReactor: func(metricName, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (*object.Metric, error) {
					return &object.Metric{
						Current: value.MetricValue{
							AverageValue: int64Ptr(17),
						},
						ReadyPodCount: int64Ptr(5),
					}, nil
				},
			},
			nil,
			nil,
			&appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"test": "test",
						},
					},
					Replicas: int32Ptr(3),
				},
			},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &config.K8sObjectMetricSource{
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},

		{
			"Single pods metric, fail to convert label",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: failed to get pods metric: "invalid" is not a valid pod selector operator`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &config.K8sPodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Selector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Operator: "invalid",
									},
								},
							},
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single pods metric, fail to get metric, non average value",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: invalid pods metric source: must be average value`),
			nil,
			nil,
			&fake.PodsGatherer{},
			nil,
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &config.K8sPodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.UtilizationMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single pods metric, fail to get metric",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: failed to get pods metric: fail to get pods metric`),
			nil,
			nil,
			&fake.PodsGatherer{
				GetMetricReactor: func(metricName, namespace string, selector, metricSelector labels.Selector) (*pods.Metric, error) {
					return nil, errors.New("fail to get pods metric")
				},
			},
			nil,
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &config.K8sPodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single pods metric, replicationcontroller, success",
			[]*k8smetric.Metric{
				{
					CurrentReplicas: 8,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.PodsMetricSourceType,
						Pods: &config.K8sPodsMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Selector: &metav1.LabelSelector{},
							},
							Target: config.K8sMetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Pods: &pods.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						ReadyPodCount:  3,
						TotalPods:      5,
						MissingPods: sets.String{
							"test-pod": {},
						},
					},
				},
			},
			nil,
			nil,
			nil,
			&fake.PodsGatherer{
				GetMetricReactor: func(metricName, namespace string, selector, metricSelector labels.Selector) (*pods.Metric, error) {
					return &pods.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						ReadyPodCount:  3,
						TotalPods:      5,
						MissingPods: sets.String{
							"test-pod": {},
						},
					}, nil
				},
			},
			nil,
			&v1.ReplicationController{
				Spec: v1.ReplicationControllerSpec{
					Selector: map[string]string{
						"test": "test",
					},
				},
			},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &config.K8sPodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 8,
				},
			},
		},
		{
			"Single resource metric, invalid target",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: invalid resource metric source: must be either average value or average utilization`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &config.K8sResourceMetricSource{
						Name: "test-resource",
						Target: config.K8sMetricTarget{
							Type: "invalid",
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single resource metric, average value, fail to get metric",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: failed to get resource metric: fail to get resource metric`),
			&fake.ResourceGatherer{
				GetRawMetricReactor: func(resource v1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error) {
					return nil, errors.New("fail to get resource metric")
				},
			},
			nil,
			nil,
			nil,
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &config.K8sResourceMetricSource{
						Name: "test-resource",
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single resource metric, average metric, success",
			[]*k8smetric.Metric{
				{
					CurrentReplicas: 9,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.ResourceMetricSourceType,
						Resource: &config.K8sResourceMetricSource{
							Name: "test-resource",
							Target: config.K8sMetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Resource: &resource.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						ReadyPodCount:  3,
						TotalPods:      5,
						MissingPods: sets.String{
							"missing-pod": {},
						},
						IgnoredPods: sets.String{
							"ignored-pod": {},
						},
						Requests: map[string]int64{
							"test-pod": 5,
						},
					},
				},
			},
			nil,
			&fake.ResourceGatherer{
				GetRawMetricReactor: func(res v1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error) {
					return &resource.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						ReadyPodCount:  3,
						TotalPods:      5,
						MissingPods: sets.String{
							"missing-pod": {},
						},
						IgnoredPods: sets.String{
							"ignored-pod": {},
						},
						Requests: map[string]int64{
							"test-pod": 5,
						},
					}, nil
				},
			},
			nil,
			nil,
			nil,
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &config.K8sResourceMetricSource{
						Name: "test-resource",
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 9,
				},
			},
		},
		{
			"Single resource metric, average utilisation, fail to get metric",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: failed to get resource metric: fail to get resource metric`),
			&fake.ResourceGatherer{
				GetMetricReactor: func(resource v1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error) {
					return nil, errors.New("fail to get resource metric")
				},
			},
			nil,
			nil,
			nil,
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &config.K8sResourceMetricSource{
						Name: "test-resource",
						Target: config.K8sMetricTarget{
							Type: autoscaling.UtilizationMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single resource metric, average utilisation, success",
			[]*k8smetric.Metric{
				{
					CurrentReplicas: 9,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.ResourceMetricSourceType,
						Resource: &config.K8sResourceMetricSource{
							Name: "test-resource",
							Target: config.K8sMetricTarget{
								Type: autoscaling.UtilizationMetricType,
							},
						},
					},
					Resource: &resource.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						ReadyPodCount:  3,
						TotalPods:      5,
						MissingPods: sets.String{
							"missing-pod": {},
						},
						IgnoredPods: sets.String{
							"ignored-pod": {},
						},
					},
				},
			},
			nil,
			&fake.ResourceGatherer{
				GetMetricReactor: func(res v1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error) {
					return &resource.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						ReadyPodCount:  3,
						TotalPods:      5,
						MissingPods: sets.String{
							"missing-pod": {},
						},
						IgnoredPods: sets.String{
							"ignored-pod": {},
						},
					}, nil
				},
			},
			nil,
			nil,
			nil,
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"test": "test",
						},
					},
				},
			},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &config.K8sResourceMetricSource{
						Name: "test-resource",
						Target: config.K8sMetricTarget{
							Type: autoscaling.UtilizationMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 9,
				},
			},
		},
		{
			"Single external metric, invalid target",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: invalid external metric source: must be either value or average value`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &config.K8sExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "test-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: "invalid",
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single external metric, average value, fail to get metric",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: failed to get external metric: fail to get metric`),
			nil,
			nil,
			nil,
			&fake.ExternalGatherer{
				GetPerPodMetricReactor: func(metricName, namespace string, metricSelector *metav1.LabelSelector) (*external.Metric, error) {
					return nil, errors.New("fail to get metric")
				},
			},
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &config.K8sExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "test-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single external metric, average metric, success",
			[]*k8smetric.Metric{
				{
					CurrentReplicas: 2,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.ExternalMetricSourceType,
						External: &config.K8sExternalMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "test-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: config.K8sMetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					External: &external.Metric{
						Current: value.MetricValue{
							AverageValue: int64Ptr(5),
						},
						ReadyPodCount: int64Ptr(6),
					},
				},
			},
			nil,
			nil,
			nil,
			nil,
			&fake.ExternalGatherer{
				GetPerPodMetricReactor: func(metricName, namespace string, metricSelector *metav1.LabelSelector) (*external.Metric, error) {
					return &external.Metric{
						Current: value.MetricValue{
							AverageValue: int64Ptr(5),
						},
						ReadyPodCount: int64Ptr(6),
					}, nil
				},
			},
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"test": "test",
						},
					},
				},
			},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &config.K8sExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "test-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 2,
				},
			},
		},
		{
			"Single external metric, value, fail to get metric",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: failed to get external metric: fail to get metric`),
			nil,
			nil,
			nil,
			&fake.ExternalGatherer{
				GetMetricReactor: func(metricName, namespace string, metricSelector *metav1.LabelSelector, podSelector labels.Selector) (*external.Metric, error) {
					return nil, errors.New("fail to get metric")
				},
			},
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &config.K8sExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "test-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.ValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"Single external metric, value, success",
			[]*k8smetric.Metric{
				{
					CurrentReplicas: 7,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.ExternalMetricSourceType,
						External: &config.K8sExternalMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "test-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: config.K8sMetricTarget{
								Type: autoscaling.ValueMetricType,
							},
						},
					},
					External: &external.Metric{
						Current: value.MetricValue{
							Value: int64Ptr(2),
						},
						ReadyPodCount: int64Ptr(3),
					},
				},
			},
			nil,
			nil,
			nil,
			nil,
			&fake.ExternalGatherer{
				GetMetricReactor: func(metricName, namespace string, metricSelector *metav1.LabelSelector, podSelector labels.Selector) (*external.Metric, error) {
					return &external.Metric{
						Current: value.MetricValue{
							Value: int64Ptr(2),
						},
						ReadyPodCount: int64Ptr(3),
					}, nil
				},
			},
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"test": "test",
						},
					},
				},
			},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &config.K8sExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "test-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.ValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 7,
				},
			},
		},
		{
			"One of each metric, all failure",
			nil,
			errors.New(`invalid metrics (4 invalid out of 4), first error is: failed to get external metric: fail to get external metric`),
			&fake.ResourceGatherer{
				GetRawMetricReactor: func(resource v1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error) {
					return nil, errors.New("fail to get resource metric")
				},
			},
			&fake.ObjectGatherer{
				GetPerPodMetricReactor: func(metricName, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (*object.Metric, error) {
					return nil, errors.New("fail to get object metric")
				},
			},
			&fake.PodsGatherer{
				GetMetricReactor: func(metricName, namespace string, selector, metricSelector labels.Selector) (*pods.Metric, error) {
					return nil, errors.New("fail to get pods metric")
				},
			},
			&fake.ExternalGatherer{
				GetMetricReactor: func(metricName, namespace string, metricSelector *metav1.LabelSelector, podSelector labels.Selector) (*external.Metric, error) {
					return nil, errors.New("fail to get external metric")
				},
			},
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &config.K8sExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "external-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.ValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &config.K8sPodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "pods-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &config.K8sObjectMetricSource{
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &config.K8sResourceMetricSource{
						Name: "test-resource",
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 3,
				},
			},
		},
		{
			"One of each metric, 2 success, 2 invalid",
			[]*k8smetric.Metric{
				{
					CurrentReplicas: 4,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.ExternalMetricSourceType,
						External: &config.K8sExternalMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "external-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: config.K8sMetricTarget{
								Type: autoscaling.ValueMetricType,
							},
						},
					},
					External: &external.Metric{
						Current: value.MetricValue{
							Value: int64Ptr(2),
						},
						ReadyPodCount: int64Ptr(3),
					},
				},
				{
					CurrentReplicas: 4,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.PodsMetricSourceType,
						Pods: &config.K8sPodsMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "pods-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: config.K8sMetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Pods: &pods.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						ReadyPodCount:  3,
						IgnoredPods: sets.String{
							"ignored-pod": {},
						},
						MissingPods: sets.String{
							"missing-pod": {},
						},
						TotalPods: 5,
					},
				},
			},
			nil,
			&fake.ResourceGatherer{
				GetRawMetricReactor: func(resource v1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error) {
					return nil, errors.New("fail to get resource metric")
				},
			},
			&fake.ObjectGatherer{
				GetPerPodMetricReactor: func(metricName, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (*object.Metric, error) {
					return nil, errors.New("fail to get object metric")
				},
			},
			&fake.PodsGatherer{
				GetMetricReactor: func(metricName, namespace string, selector, metricSelector labels.Selector) (*pods.Metric, error) {
					return &pods.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						ReadyPodCount:  3,
						IgnoredPods: sets.String{
							"ignored-pod": {},
						},
						MissingPods: sets.String{
							"missing-pod": {},
						},
						TotalPods: 5,
					}, nil
				},
			},
			&fake.ExternalGatherer{
				GetMetricReactor: func(metricName, namespace string, metricSelector *metav1.LabelSelector, podSelector labels.Selector) (*external.Metric, error) {
					return &external.Metric{
						Current: value.MetricValue{
							Value: int64Ptr(2),
						},
						ReadyPodCount: int64Ptr(3),
					}, nil
				},
			},
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &config.K8sExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "external-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.ValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &config.K8sPodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "pods-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &config.K8sObjectMetricSource{
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &config.K8sResourceMetricSource{
						Name: "test-resource",
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 4,
				},
			},
		},
		{
			"One of each metric, all success",
			[]*k8smetric.Metric{
				{
					CurrentReplicas: 4,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.ExternalMetricSourceType,
						External: &config.K8sExternalMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "external-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: config.K8sMetricTarget{
								Type: autoscaling.ValueMetricType,
							},
						},
					},
					External: &external.Metric{
						Current: value.MetricValue{
							Value: int64Ptr(2),
						},
						ReadyPodCount: int64Ptr(3),
					},
				},
				{
					CurrentReplicas: 4,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.PodsMetricSourceType,
						Pods: &config.K8sPodsMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "pods-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: config.K8sMetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Pods: &pods.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						ReadyPodCount:  3,
						IgnoredPods: sets.String{
							"ignored-pod": {},
						},
						MissingPods: sets.String{
							"missing-pod": {},
						},
						TotalPods: 5,
					},
				},
				{
					CurrentReplicas: 4,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.ObjectMetricSourceType,
						Object: &config.K8sObjectMetricSource{
							Target: config.K8sMetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Object: &object.Metric{
						Current: value.MetricValue{
							AverageValue: int64Ptr(2),
						},
						ReadyPodCount: int64Ptr(6),
					},
				},
				{
					CurrentReplicas: 4,
					Spec: config.K8sMetricSpec{
						Type: autoscaling.ResourceMetricSourceType,
						Resource: &config.K8sResourceMetricSource{
							Name: "test-resource",
							Target: config.K8sMetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Resource: &resource.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						Requests:       map[string]int64{"pod-1": 1, "pod-2": 3, "pod-3": 4},
						ReadyPodCount:  4,
						TotalPods:      6,
						IgnoredPods:    sets.String{"pod-1": {}},
						MissingPods:    sets.String{"pod-3": {}},
					},
				},
			},
			nil,
			&fake.ResourceGatherer{
				GetRawMetricReactor: func(res v1.ResourceName, namespace string, selector labels.Selector) (*resource.Metric, error) {
					return &resource.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						Requests:       map[string]int64{"pod-1": 1, "pod-2": 3, "pod-3": 4},
						ReadyPodCount:  4,
						TotalPods:      6,
						IgnoredPods:    sets.String{"pod-1": {}},
						MissingPods:    sets.String{"pod-3": {}},
					}, nil
				},
			},
			&fake.ObjectGatherer{
				GetPerPodMetricReactor: func(metricName, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (*object.Metric, error) {
					return &object.Metric{
						Current: value.MetricValue{
							AverageValue: int64Ptr(2),
						},
						ReadyPodCount: int64Ptr(6),
					}, nil
				},
			},
			&fake.PodsGatherer{
				GetMetricReactor: func(metricName, namespace string, selector, metricSelector labels.Selector) (*pods.Metric, error) {
					return &pods.Metric{
						PodMetricsInfo: podmetrics.MetricsInfo{},
						ReadyPodCount:  3,
						IgnoredPods: sets.String{
							"ignored-pod": {},
						},
						MissingPods: sets.String{
							"missing-pod": {},
						},
						TotalPods: 5,
					}, nil
				},
			},
			&fake.ExternalGatherer{
				GetMetricReactor: func(metricName, namespace string, metricSelector *metav1.LabelSelector, podSelector labels.Selector) (*external.Metric, error) {
					return &external.Metric{
						Current: value.MetricValue{
							Value: int64Ptr(2),
						},
						ReadyPodCount: int64Ptr(3),
					}, nil
				},
			},
			&appsv1.Deployment{},
			[]config.K8sMetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &config.K8sExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "external-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.ValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &config.K8sPodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "pods-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &config.K8sObjectMetricSource{
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &config.K8sResourceMetricSource{
						Name: "test-resource",
						Target: config.K8sMetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
			&autoscalingv1.Scale{
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 4,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := k8smetricget.Gather{
				Resource: test.resource,
				Pods:     test.pods,
				Object:   test.object,
				External: test.external,
			}
			metrics, err := gatherer.GetMetrics(test.deployment, test.specs, test.namespace, test.scaleResource)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, metrics) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, metrics))
			}
		})
	}
}

func TestNewGather(t *testing.T) {
	var tests = []struct {
		description                   string
		expected                      k8smetricget.Gatherer
		metricsClient                 metricsclient.Client
		podlister                     corelisters.PodLister
		cpuInitializationPeriod       time.Duration
		delayOfInitialReadinessStatus time.Duration
	}{
		{
			"Set up all sub gatherers",
			&k8smetricget.Gather{
				Resource: &resourceget.Gather{
					MetricsClient:                 &fake.MetricClient{},
					PodLister:                     &fake.PodLister{},
					CPUInitializationPeriod:       2,
					DelayOfInitialReadinessStatus: 2,
				},
				Pods: &podsget.Gather{
					MetricsClient: &fake.MetricClient{},
					PodLister:     &fake.PodLister{},
				},
				Object: &objectget.Gather{
					MetricsClient: &fake.MetricClient{},
					PodReadyCounter: &podutil.PodReadyCount{
						PodLister: &fake.PodLister{},
					},
				},
				External: &externalget.Gather{
					MetricsClient: &fake.MetricClient{},
					PodReadyCounter: &podutil.PodReadyCount{
						PodLister: &fake.PodLister{},
					},
				},
			},
			&fake.MetricClient{},
			&fake.PodLister{},
			2,
			2,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := k8smetricget.NewGather(test.metricsClient, test.podlister, test.cpuInitializationPeriod, test.delayOfInitialReadinessStatus)
			if !cmp.Equal(test.expected, gatherer) {
				t.Errorf("gatherer mismatch (-want +got):\n%s", cmp.Diff(test.expected, gatherer))
			}
		})
	}
}
