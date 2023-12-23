module github.com/jthomperoo/custom-pod-autoscaler/example/simple-pod-metrics-golang

go 1.16

require (
	github.com/jthomperoo/custom-pod-autoscaler/v2 v2.9.0
	k8s.io/api v0.21.11
	k8s.io/apimachinery v0.21.11
	k8s.io/client-go v0.21.11
)

replace github.com/jthomperoo/custom-pod-autoscaler/v2 => ../../
