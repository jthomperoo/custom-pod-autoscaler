# Golang Simple Pod Metrics Example
This example shows how to make a Custom Pod Autoscaler (CPA) using Golang.
The example extends the Alpine CPA base image (custompodautoscaler/alpine) and adds in a Golang binary to execute for gathering metrics/making evaluations.
The code is verbosely commented and designed to be read and understood for building your own CPAs.

## Overview
This example contains a docker image of the example Golang Custom Pod Autoscaler, alongside using the `flask-metric` sample application as a target to scale up and down.

### Example Custom Pod Autoscaler

The CPA will try to ensure that there is always atleast `1` available across the resource and each app in the resource.
The CPA will also ensure there are no more than `5` available across the resource.
The CPA exposes two endpoints:
* `GET /metrics`
    * Displays all gathered metric values from every app instance in the resource.
    * Aliased with `metrics` in the example Dockerfile.
* `GET /evaluation`
    * Displays the evaluation decision made on how to scale, reporting the `targetReplicas` - how many replicas the resource should have.
    * Aliased with `evaluation` in the example Dockerfile.

## Usage
Trying out this example requires a kubernetes cluster to try it out on, this guide will assume you are using Minikube.

### Enable CPAs
Using this CPA requires CPAs to be enabled on your kubernetes cluster, [follow this guide to set up CPAs on your cluster](https://github.com/jthomperoo/custom-pod-autoscaler-operator#installation).

### Switch to target the Minikube registry
Target the Minikube registry for building the image:
`eval $(minikube docker-env)`

### Deploy an app for the CPA to manage
You need to deploy an app for the CPA to manage:
* Build the example app image.
`docker build -t flask-metric ../flask-metric`
* Deploy the app using a deployment.
`kubectl apply -f ../flask-metric/deployment.yaml`
Now you have an app running to manage scaling for.

### Build CPA image
Once CPAs have been enabled on your cluster, you need to build this example, run these commands to build the example:
* Use the Makefile to build the docker image
`make docker`
* Deploy the CPA using the image just built.
`kubectl apply -f cpa.yaml`
Now the CPA should be running on your cluster, managing the app we previously deployed.

## Testing the CPA
* List pods.
`kubectl get pods -l app=flask-metric`
* Exec into a pod.
`kubectl exec -it POD_NAME bash`
* Get value.
`metric`
* Increment value.
`increment`
* Decrement value.
`decrement`

## Using this example as a starting point

This example uses a `go.mod` file to manage dependencies, you will need to set up a new `go.mod` file which uses
your own package name. This example also uses a `go.mod` import that relies directly on the latest version of the
custom pod autoscaler Go code, for your own autoscaler you would need to run:

```bash
go get github.com/jthomperoo/custom-pod-autoscaler/v2
```
