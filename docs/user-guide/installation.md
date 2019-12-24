# Installation
The easiest way to install a Custom Pod Autoscaler is using the 
[Custom Pod Autoscaler Operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator), 
follow the 
[installation guide for instructions for installing the operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/master/INSTALL.md).

Once the operator is installed, you can now use the new `CustomPodAutoscaler` Kubernetes Resource 
to deploy autoscalers.  

For an example, taken from the 
[getting started guide](../getting-started) 
and the 
[example code](https://github.com/jthomperoo/custom-pod-autoscaler/tree/master/example/python-custom-autoscaler)
in a file called `cpa.yaml`:
```yaml
apiVersion: custompodautoscaler.com/v1alpha1
kind: CustomPodAutoscaler
metadata:
  name: python-custom-autoscaler
spec:
  template:
    spec:
      containers:
      - name: python-custom-autoscaler
        image: python-custom-autoscaler:latest
        imagePullPolicy: IfNotPresent
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: hello-kubernetes
  config: 
    - name: interval
      value: "10000"
```
We can deploy this using `kubectl apply -f cpa.yaml`.