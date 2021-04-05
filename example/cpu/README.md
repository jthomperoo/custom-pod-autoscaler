# K8s CPU Metrics example

This example shows how to make a Custom Pod Autoscaler (CPA) that uses metrics fetched from the Kubernetes (K8s) metrics
server.
This example uses CPU metrics for autoscaling.
The code is verbosely commented and designed to be read and understood for building your own CPAs.

## Overview

This example contains a docker image of the example Python Custom Pod Autoscaler; targeting the `php-apache` deployment

### Example Custom Pod Autoscaler


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
