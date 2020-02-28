# Downscale Stabilization

This is an autoscaler that makes use of the `downscaleStabilization` configuration option. This is based on the [simple pod metrics python example](../simple-pod-metrics-python), augmented with the `downscaleStabilization` option.

## Overview

This autoscaler is functionally the same as the [simple pod metrics python example](../simple-pod-metrics-python), but with the `downscaleStabilization` option.  

The `downscaleStabilization` option is defined in the `cpa.yaml` file as a runtime configuration option:

```yaml
  config: 
    - name: interval
      value: "10000"
    - name: downscaleStabilization
      value: "60"
```

This will result in a downscale stabilization window of 60 seconds, during which time it will always pick the evaluation with the highest replica count.  

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
* Build the example image.  
`docker build -t downscale-stabilization .`  
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

## Conclusion

Have a look at [the wiki for more information, such as guides and references](https://custom-pod-autoscaler.readthedocs.io/en/latest/)