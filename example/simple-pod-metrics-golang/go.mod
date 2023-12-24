module github.com/jthomperoo/custom-pod-autoscaler/example/simple-pod-metrics-golang

go 1.16

require (
	github.com/jthomperoo/custom-pod-autoscaler/v2 v2.9.0
	k8s.io/api v0.29.0
	k8s.io/apimachinery v0.29.0
	k8s.io/client-go v0.29.0
)

replace github.com/jthomperoo/custom-pod-autoscaler/v2 => ../../
