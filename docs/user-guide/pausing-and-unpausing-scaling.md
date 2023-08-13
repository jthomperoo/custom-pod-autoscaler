# Pausing and Un-pausing Scaling

If you are using the Custom Pod Autoscaler Operator version `v1.4.0` and above you can pause and unpause scaling
on an autoscaler.

If you want to disable an autoscaler from autoscaling (e.g. during maintenance) you can do so by using the
`v1.custompodautoscaler.com/paused-replicas` annotation on the Custom Pod Autoscaler.

See the [Custom Pod Autoscaler Operator usage guide for more
details](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/v1.4.0/USAGE.md#pausing-autoscaling).
