# Python Custom Pod Autoscaler
This example shows how to make a Custom Pod Autoscaler (CPA) using Python. The example extends the Python CPA base image (custompodautoscaler/python) and sets up an environment to allow python scripts to be used to determine metrics and evaluate how to scale. The code is verbosely commented and designed to be read and understood for building your own CPAs.

## Overview
This example contains two docker images, one for the CPA and one for an app that the CPA will manage.  
The CPA scales based on a custom endpoint exposed by the app, `GET /metrics` - this returns a JSON metric with an arbitrary value between `0` and `5`.  
Example metric:  
```
{
    "value": 3,
    "available": 2,
    "min": 0,
    "max": 5
}
```  
This value can be incremented and decremented with `POST /increment` and `POST /decrement` respectively. Available is how far the value is from the max.
The CPA will try to ensure that there is always atleast `1` available across the deployment and each pod in the deployment. The CPA will also ensure there are no more than `5` available across the deployment. You can exec into the example app pod and increase/decrease the value and see how the CPA creates/deletes pods.  

## Usage
Trying out this example requires a kubernetes cluster to try it out on, this guide will assume you are using Minikube.  

### Enable CPAs
Using this CPA requires CPAs to be enabled on your kubernetes cluster, [follow this guide to set up CPAs on your cluster](https://github.com/jthomperoo/custom-pod-autoscaler-operator#installation).  

### Build CPA image
Once CPAs have been enabled on your cluster, you need to build this example, run these commands to build the example:  
* Target the Minikube registry for building the image.  
`eval $(minikube docker-env)`  
* Build the example image.  
`docker build -t example-python-custom-pod-autoscaler .`  
* Deploy the CPA using the image just built.  
`kubectl apply -f cpa.yaml`  

### Deploy a deployment for the CPA to manage
Now you have a CPA running on the cluster, you need to deploy an app for it to manage:  
* Target the Minikube registry for building the image.  
`eval $(minikube docker-env)`  
* Move to the app directory.  
`cd app`  
* Build the example app image.  
`docker build -t flask-metric .`  
* Deploy the app using a deployment.  
`kubectl apply -f deployment.yaml`  
Now you have CPAs enabled on your cluster, the example CPA running and an app deployment that it is managing.

## Testing the CPA
* List pods.  
`kubectl get pods -l app=flask-metric`  
* Exec into a pod.  
`kubectl exec -it POD_NAME bash`  
* Get value.  
`curl http://localhost:5000/metric`  
* Increment value.  
`curl -X POST http://localhost:5000/increment`  
* Decrement value.  
`curl -X POST http://localhost:5000/decrement`