package external

import (
	"time"

	"github.com/jthomperoo/custom-pod-autoscaler/k8smetric/value"
)

// Metric (Resource) is a global metric that is not associated with any Kubernetes object. It allows autoscaling based
// on information coming from components running outside of cluster (for example length of queue in cloud messaging
// service, or QPS from loadbalancer running outside of cluster).
type Metric struct {
	Current       value.MetricValue `json:"current,omitempty"`
	ReadyPodCount *int64            `json:"ready_pod_count,omitempty"`
	Timestamp     time.Time         `json:"timestamp,omitempty"`
}
