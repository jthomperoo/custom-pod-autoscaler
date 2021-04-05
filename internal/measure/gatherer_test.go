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

package measure_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/external"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/object"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/pods"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/podutil"
	"github.com/jthomperoo/custom-pod-autoscaler/internal/measure/resource"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	corelisters "k8s.io/client-go/listers/core/v1"
	metricsclient "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
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
		description string
		expected    []*measure.Metric
		expectedErr error
		resource    resource.Gatherer
		object      object.Gatherer
		pods        pods.Gatherer
		external    external.Gatherer
		deployment  metav1.Object
		specs       []measure.MetricSpec
		namespace   string
	}{
		{
			"Single invalid resource type",
			nil,
			errors.New(`Unsupported resource of type *v1.DaemonSet`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.DaemonSet{},
			[]measure.MetricSpec{
				{
					Type: "invalid",
				},
			},
			"test-namespace",
		},
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
			[]measure.MetricSpec{
				{
					Type: "invalid",
				},
			},
			"test-namespace",
		},
		{
			"Single object metric, fail to convert label",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: failed to get object metric: "invalid" is not a valid pod selector operator`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &measure.ObjectMetricSource{
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
		},
		{
			"Single object metric, invalid target",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: invalid object metric source: neither a value target nor an average value target was set`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &measure.ObjectMetricSource{
						Target: measure.MetricTarget{
							Type: "invalid",
						},
					},
				},
			},
			"test-namespace",
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
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &measure.ObjectMetricSource{
						Target: measure.MetricTarget{
							Type: autoscaling.ValueMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"Single object metric, deployment, value metric, success",
			[]*measure.Metric{
				{
					CurrentReplicas: 1,
					Spec: measure.MetricSpec{
						Type: autoscaling.ObjectMetricSourceType,
						Object: &measure.ObjectMetricSource{
							Target: measure.MetricTarget{
								Type: autoscaling.ValueMetricType,
							},
						},
					},
					Object: &object.Metric{
						Utilization:   5,
						ReadyPodCount: int64Ptr(2),
					},
				},
			},
			nil,
			nil,
			&fake.ObjectGatherer{
				GetMetricReactor: func(metricName, namespace string, objectRef *autoscaling.CrossVersionObjectReference, selector, metricSelector labels.Selector) (*object.Metric, error) {
					return &object.Metric{
						Utilization:   5,
						ReadyPodCount: int64Ptr(2),
					}, nil
				},
			},
			nil,
			nil,
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &measure.ObjectMetricSource{
						Target: measure.MetricTarget{
							Type: autoscaling.ValueMetricType,
						},
					},
				},
			},
			"test-namespace",
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
			&appsv1.ReplicaSet{
				Spec: appsv1.ReplicaSetSpec{
					Replicas: int32Ptr(1),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &measure.ObjectMetricSource{
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"Single object metric, statefulset, average value metric, success",
			[]*measure.Metric{
				{
					CurrentReplicas: 3,
					Spec: measure.MetricSpec{
						Type: autoscaling.ObjectMetricSourceType,
						Object: &measure.ObjectMetricSource{
							Target: measure.MetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Object: &object.Metric{
						Utilization:   17,
						ReadyPodCount: int64Ptr(5),
					},
				},
			},
			nil,
			nil,
			&fake.ObjectGatherer{
				GetPerPodMetricReactor: func(metricName, namespace string, objectRef *autoscaling.CrossVersionObjectReference, metricSelector labels.Selector) (*object.Metric, error) {
					return &object.Metric{
						Utilization:   17,
						ReadyPodCount: int64Ptr(5),
					}, nil
				},
			},
			nil,
			nil,
			&appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: int32Ptr(3),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &measure.ObjectMetricSource{
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"Single pods metric, fail to convert label",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: failed to get pods metric: "invalid" is not a valid pod selector operator`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &measure.PodsMetricSource{
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
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &measure.PodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Selector: &metav1.LabelSelector{},
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"Single pods metric, replicationcontroller, success",
			[]*measure.Metric{
				{
					CurrentReplicas: 8,
					Spec: measure.MetricSpec{
						Type: autoscaling.PodsMetricSourceType,
						Pods: &measure.PodsMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Selector: &metav1.LabelSelector{},
							},
						},
					},
					Pods: &pods.Metric{
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
					Replicas: int32Ptr(8),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &measure.PodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Selector: &metav1.LabelSelector{},
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"Single resource metric, invalid target",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: invalid resource metric source: must be either average value or average utilization`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(3),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &measure.ResourceMetricSource{
						Name: "test-resource",
						Target: measure.MetricTarget{
							Type: "invalid",
						},
					},
				},
			},
			"test-namespace",
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
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &measure.ResourceMetricSource{
						Name: "test-resource",
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"Single resource metric, average metric, success",
			[]*measure.Metric{
				{
					CurrentReplicas: 9,
					Spec: measure.MetricSpec{
						Type: autoscaling.ResourceMetricSourceType,
						Resource: &measure.ResourceMetricSource{
							Name: "test-resource",
							Target: measure.MetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Resource: &resource.Metric{
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(9),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &measure.ResourceMetricSource{
						Name: "test-resource",
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
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
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &measure.ResourceMetricSource{
						Name: "test-resource",
						Target: measure.MetricTarget{
							Type: autoscaling.UtilizationMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"Single resource metric, average utilisation, success",
			[]*measure.Metric{
				{
					CurrentReplicas: 9,
					Spec: measure.MetricSpec{
						Type: autoscaling.ResourceMetricSourceType,
						Resource: &measure.ResourceMetricSource{
							Name: "test-resource",
							Target: measure.MetricTarget{
								Type: autoscaling.UtilizationMetricType,
							},
						},
					},
					Resource: &resource.Metric{
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
					Replicas: int32Ptr(9),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &measure.ResourceMetricSource{
						Name: "test-resource",
						Target: measure.MetricTarget{
							Type: autoscaling.UtilizationMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"Single external metric, invalid target",
			nil,
			errors.New(`invalid metrics (1 invalid out of 1), first error is: invalid external metric source: must be either average value or average utilization`),
			nil,
			nil,
			nil,
			nil,
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(3),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &measure.ExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "test-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: measure.MetricTarget{
							Type: "invalid",
						},
					},
				},
			},
			"test-namespace",
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
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(2),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &measure.ExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "test-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"Single external metric, average metric, success",
			[]*measure.Metric{
				{
					CurrentReplicas: 2,
					Spec: measure.MetricSpec{
						Type: autoscaling.ExternalMetricSourceType,
						External: &measure.ExternalMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "test-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: measure.MetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					External: &external.Metric{
						Utilization:   5,
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
						Utilization:   5,
						ReadyPodCount: int64Ptr(6),
					}, nil
				},
			},
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(2),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &measure.ExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "test-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"Single external metric, average utilisation, fail to get metric",
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
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(7),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &measure.ExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "test-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: measure.MetricTarget{
							Type: autoscaling.UtilizationMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"Single external metric, average utilisation, success",
			[]*measure.Metric{
				{
					CurrentReplicas: 7,
					Spec: measure.MetricSpec{
						Type: autoscaling.ExternalMetricSourceType,
						External: &measure.ExternalMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "test-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: measure.MetricTarget{
								Type: autoscaling.UtilizationMetricType,
							},
						},
					},
					External: &external.Metric{
						Utilization:   2,
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
						Utilization:   2,
						ReadyPodCount: int64Ptr(3),
					}, nil
				},
			},
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(7),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &measure.ExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "test-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: measure.MetricTarget{
							Type: autoscaling.UtilizationMetricType,
						},
					},
				},
			},
			"test-namespace",
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
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(4),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &measure.ExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "external-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: measure.MetricTarget{
							Type: autoscaling.UtilizationMetricType,
						},
					},
				},
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &measure.PodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "pods-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &measure.ObjectMetricSource{
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &measure.ResourceMetricSource{
						Name: "test-resource",
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"One of each metric, 2 success, 2 invalid",
			[]*measure.Metric{
				{
					CurrentReplicas: 4,
					Spec: measure.MetricSpec{
						Type: autoscaling.ExternalMetricSourceType,
						External: &measure.ExternalMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "external-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: measure.MetricTarget{
								Type: autoscaling.UtilizationMetricType,
							},
						},
					},
					External: &external.Metric{
						Utilization:   2,
						ReadyPodCount: int64Ptr(3),
					},
				},
				{
					CurrentReplicas: 4,
					Spec: measure.MetricSpec{
						Type: autoscaling.PodsMetricSourceType,
						Pods: &measure.PodsMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "pods-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: measure.MetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Pods: &pods.Metric{
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
						Utilization:   2,
						ReadyPodCount: int64Ptr(3),
					}, nil
				},
			},
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(4),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &measure.ExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "external-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: measure.MetricTarget{
							Type: autoscaling.UtilizationMetricType,
						},
					},
				},
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &measure.PodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "pods-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &measure.ObjectMetricSource{
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &measure.ResourceMetricSource{
						Name: "test-resource",
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
		{
			"One of each metric, all success",
			[]*measure.Metric{
				{
					CurrentReplicas: 4,
					Spec: measure.MetricSpec{
						Type: autoscaling.ExternalMetricSourceType,
						External: &measure.ExternalMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "external-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: measure.MetricTarget{
								Type: autoscaling.UtilizationMetricType,
							},
						},
					},
					External: &external.Metric{
						Utilization:   2,
						ReadyPodCount: int64Ptr(3),
					},
				},
				{
					CurrentReplicas: 4,
					Spec: measure.MetricSpec{
						Type: autoscaling.PodsMetricSourceType,
						Pods: &measure.PodsMetricSource{
							Metric: autoscaling.MetricIdentifier{
								Name:     "pods-metric",
								Selector: &metav1.LabelSelector{},
							},
							Target: measure.MetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Pods: &pods.Metric{
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
					Spec: measure.MetricSpec{
						Type: autoscaling.ObjectMetricSourceType,
						Object: &measure.ObjectMetricSource{
							Target: measure.MetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Object: &object.Metric{
						Utilization:   2,
						ReadyPodCount: int64Ptr(6),
					},
				},
				{
					CurrentReplicas: 4,
					Spec: measure.MetricSpec{
						Type: autoscaling.ResourceMetricSourceType,
						Resource: &measure.ResourceMetricSource{
							Name: "test-resource",
							Target: measure.MetricTarget{
								Type: autoscaling.AverageValueMetricType,
							},
						},
					},
					Resource: &resource.Metric{
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
						Utilization:   2,
						ReadyPodCount: int64Ptr(6),
					}, nil
				},
			},
			&fake.PodsGatherer{
				GetMetricReactor: func(metricName, namespace string, selector, metricSelector labels.Selector) (*pods.Metric, error) {
					return &pods.Metric{
						PodMetricsInfo: metricsclient.PodMetricsInfo{},
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
						Utilization:   2,
						ReadyPodCount: int64Ptr(3),
					}, nil
				},
			},
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(4),
				},
			},
			[]measure.MetricSpec{
				{
					Type: autoscaling.ExternalMetricSourceType,
					External: &measure.ExternalMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "external-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: measure.MetricTarget{
							Type: autoscaling.UtilizationMetricType,
						},
					},
				},
				{
					Type: autoscaling.PodsMetricSourceType,
					Pods: &measure.PodsMetricSource{
						Metric: autoscaling.MetricIdentifier{
							Name:     "pods-metric",
							Selector: &metav1.LabelSelector{},
						},
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ObjectMetricSourceType,
					Object: &measure.ObjectMetricSource{
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
				{
					Type: autoscaling.ResourceMetricSourceType,
					Resource: &measure.ResourceMetricSource{
						Name: "test-resource",
						Target: measure.MetricTarget{
							Type: autoscaling.AverageValueMetricType,
						},
					},
				},
			},
			"test-namespace",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gatherer := measure.Gather{
				Resource: test.resource,
				Pods:     test.pods,
				Object:   test.object,
				External: test.external,
			}
			metrics, err := gatherer.GetMetrics(test.deployment, test.specs, test.namespace)
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
		expected                      measure.Gatherer
		metricsClient                 metricsclient.MetricsClient
		podlister                     corelisters.PodLister
		cpuInitializationPeriod       time.Duration
		delayOfInitialReadinessStatus time.Duration
	}{
		{
			"Set up all sub gatherers",
			&measure.Gather{
				Resource: &resource.Gather{
					MetricsClient:                 &fake.MetricClient{},
					PodLister:                     &fake.PodLister{},
					CPUInitializationPeriod:       2,
					DelayOfInitialReadinessStatus: 2,
				},
				Pods: &pods.Gather{
					MetricsClient: &fake.MetricClient{},
					PodLister:     &fake.PodLister{},
				},
				Object: &object.Gather{
					MetricsClient: &fake.MetricClient{},
					PodReadyCounter: &podutil.PodReadyCount{
						PodLister: &fake.PodLister{},
					},
				},
				External: &external.Gather{
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
			gatherer := measure.NewGather(test.metricsClient, test.podlister, test.cpuInitializationPeriod, test.delayOfInitialReadinessStatus)
			if !cmp.Equal(test.expected, gatherer) {
				t.Errorf("gatherer mismatch (-want +got):\n%s", cmp.Diff(test.expected, gatherer))
			}
		})
	}
}
