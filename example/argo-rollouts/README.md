# Argo Rollouts

> Note: this feature is only available in the Custom Pod Autoscaler `v2.3.0` and above.

> Note: this example requires using a Custom Pod Autoscaler Operator `v1.2.0` and above.

This is an autoscaler that targets an [Argo Rollout](https://argoproj.github.io/argo-rollouts/). This is based on the [getting started tutorial example autoscaler](../python-custom-autoscaler).

## Overview

This example sets up an Argo Rollout, initially configured to use `2` replicas. The Argo Rollout has a label called
`numPods` on it that this autoscaler will read and scale to the number provided in this label. For example if
`numPods: 5` is provided the autoscaler will read this value and scale the Rollout up to `5` replicas.

## Usage

The following steps are written with the understanding that you are using a [k3d cluster](https://k3d.io/stable/), if
you are using something else the steps will be the same except for importing the CPA images into your image registry.

### Enable Argo Rollouts

Using this requires Argo Rollouts to be enabled on your Kubernetes cluster, [follow this guide to set up Argo Rollouts
on your cluster](https://argoproj.github.io/argo-rollouts/installation/)

### Enable CPAs

Using this CPA requires CPAs to be enabled on your Kubernetes cluster, [follow this guide to set up CPAs on your
cluster](https://github.com/jthomperoo/custom-pod-autoscaler-operator#installation).

### Deploy an Argo Rollout to Manage

First a rollout needs to be deployed that the CPA can manage, you can deploy with the following command:

```bash
kubectl apply -f rollout.yaml
```

You can check if the rollout is deployed by running this command:

```bash
kubectl argo rollouts get rollout rollouts-demo
```

You should see that the rollout is set up to initially have only `2` replicas.

### Build and deploy CPA image

Next build the autoscaler:

```bash
docker build -t argo-rollouts .
```

Import the autoscaler to the k3d image registry (if using a different K8s provider this step will be different):

```bash
k3d image import argo-rollouts:latest
```

Deploy the autoscaler:

```bash
kubectl apply -f cpa.yaml
```

You can check if the rollout is being managed by running this command:

```bash
kubectl argo rollouts get rollout rollouts-demo
```

You should see that the rollout has now been scaled up to `5` replicas, based on the `numPods` label - try adjusting
this label and see the scaling occur.
