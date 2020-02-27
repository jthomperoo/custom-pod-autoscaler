module github.com/jthomperoo/custom-pod-autoscaler/example/simple-pod-metrics-golang

go 1.13

require (
	github.com/jthomperoo/custom-pod-autoscaler v0.0.0
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
)

replace github.com/jthomperoo/custom-pod-autoscaler => ../../
