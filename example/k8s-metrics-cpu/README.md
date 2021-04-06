# K8s Metrics CPU example

> Note: This example makes use of the `roleRequiresMetricsServer: true` flag of the Custom Pod Autoscaler Operator,
> this feature is *only available in Custom Pod Autoscaler Operator `v1.1.0` and above*

This example shows how to make a Custom Pod Autoscaler (CPA) that uses metrics fetched from the Kubernetes (K8s) metrics
server.
This example uses CPU metrics for autoscaling.
The code is verbosely commented and designed to be read and understood for building your own CPAs.

## Overview

This example contains a docker image of the example Python Custom Pod Autoscaler; targeting the `php-apache` deployment

### Example Custom Pod Autoscaler

The `config.yaml` file contains some configuration that sets up the Custom Pod Autoscaler to query the Kubernetes
metrics server to fetch CPU utilization values:

```yaml
kubernetesMetricSpecs:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
requireKubernetesMetrics: true
```

The above configuration snippet defines the following:

- `kubernetesMetricSpecs` is a list of metric specs, [check out the
wiki](https://custom-pod-autoscaler.readthedocs.io/en/latest/reference/configuration/#kubernetesmetricspecs) and the
[Horizontal Pod Autoscaler
docs](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/#autoscaling-on-multiple-metrics-and-custom-metrics)
for more detailed info.
    - `type: Resource` is the type of the metric spec, it is a resource metric spec.
    - `resource` is the configuration for the resource metric spec
        - `name: cpu` is targeting the CPU metric
        - `target` defines the targeting type of the metric spec
            - `type: Utilization` results are targeting CPU utilization, rather than raw CPU values.
- `requireKubernetesMetrics:true` is a flag that means that the autoscaler will fail and not call the metric gathering
if the metrics server is not available/there is failure querying the metrics server. This helps to diagnose issues
instead of causing issues in the metric server.

## Usage

Trying out this example requires a kubernetes cluster to try it out on, this guide will assume you are using Minikube.

### Enable CPAs

Using this CPA requires CPAs to be enabled on your kubernetes cluster, [follow this guide to set up CPAs on your
cluster](https://github.com/jthomperoo/custom-pod-autoscaler-operator#installation).

### Switch to target the Minikube registry

Target the Minikube registry for building the image:
`eval $(minikube docker-env)`

### Deploy an app for the CPA to manage

You need to deploy an app for the CPA to manage:
* Deploy the app using a deployment.
`kubectl apply -f deployment.yaml`
Now you have an app running to manage scaling for.

### Build CPA image

Once CPAs have been enabled on your cluster, you need to build this example, run these commands to build the example:
* Build the example image.
`docker build -t cpu-scaler .`
* Deploy the CPA using the image just built.
`kubectl apply -f cpa.yaml`
Now the CPA should be running on your cluster, managing the app we previously deployed.
