package fake

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// PodReadyCounter (fake) provides a way to insert functionality into a PodReadyCounter
type PodReadyCounter struct {
	GetReadyPodsCountReactor func(namespace string, selector labels.Selector) (int64, error)
}

// GetReadyPodsCount calls the fake PodReadyCounter function
func (f *PodReadyCounter) GetReadyPodsCount(namespace string, selector labels.Selector) (int64, error) {
	return f.GetReadyPodsCountReactor(namespace, selector)
}

// PodLister (fake) provides a way to insert functionality into a PodLister
type PodLister struct {
	ListReactor func(selector labels.Selector) (ret []*corev1.Pod, err error)
	PodsReactor func(namespace string) corelisters.PodNamespaceLister
}

// List calls the fake PodLister function
func (f *PodLister) List(selector labels.Selector) (ret []*corev1.Pod, err error) {
	return f.ListReactor(selector)
}

// Pods calls the fake PodLister function
func (f *PodLister) Pods(namespace string) corelisters.PodNamespaceLister {
	return f.PodsReactor(namespace)
}

// PodNamespaceLister (fake) provides a way to insert functionality into a PodNamespaceLister
type PodNamespaceLister struct {
	ListReactor func(selector labels.Selector) (ret []*corev1.Pod, err error)
	GetReactor  func(name string) (*corev1.Pod, error)
}

// List calls the fake PodNamespaceLister function
func (f *PodNamespaceLister) List(selector labels.Selector) (ret []*corev1.Pod, err error) {
	return f.ListReactor(selector)
}

// Get calls the fake PodNamespaceLister function
func (f *PodNamespaceLister) Get(name string) (*corev1.Pod, error) {
	return f.GetReactor(name)
}
