# K8s Metrics CPU Match Expressions example

> Note: This example makes use of the `roleRequiresMetricsServer: true` flag of the Custom Pod Autoscaler Operator,
> this feature is *only available in Custom Pod Autoscaler Operator `v1.1.0` and above*

This example is exactly the same as the [K8s Metrics CPU example](./k8s-metrics-cpu) except that this example uses
a Kubernetes [match
expression](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#resources-that-support-set-based-requirements)
to match Pods and to gather their Kubernetes metrics.

## Overview

The Deployment in this example uses the following match expression:

```yaml
selector:
  matchExpressions:
    - {key: run, operator: In, values: [php-apache]}
```

This matches any Pod that has the `run: php-apache` label. The CPA uses this to specify which K8s metrics to gather.

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

### Build and deploy CPA image

Once CPAs have been enabled on your cluster, you need to build this example, run these commands to build the example:
* Build the example image.
`docker build -t k8s-metrics-cpu-match-expressions .`
* Deploy the CPA using the image just built.
`kubectl apply -f cpa.yaml`
Now the CPA should be running on your cluster, managing the app we previously deployed.

### Increase CPU load

Increase the CPU load with:

```bash
kubectl run -it --rm load-generator --image=busybox -- /bin/sh
```

Once it has loaded, run this command to create load :

```bash
while true; do wget -q -O- http://php-apache.default.svc.cluster.local; done
```

Watch as the number of replicas increases.

The autoscaler image contains some debug aliases, you can do:

```bash
kubectl exec -it k8s-metrics-cpu -- bash
```

Once you are inside the autoscaler, you can execute the following aliases:

- `metrics` - calls the metric gathering stage through the API, will display the gathered average CPU utilization value.
- `evaluation` - forces an evaluation to be calculated through the API.
