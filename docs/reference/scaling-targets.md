# Scaling targets

Custom Pod Autoscalers can target any resource that the Horizontal Pod Autoscaler can target:

* Deployments
* ReplicaSets
* StatefulSets
* ReplicationControllers

## Scale target reference

To tell a Custom Pod Autoscaler which resource to target, provide a `scaleTargetRef` - a description of the resource to target. Within a Custom Pod Autoscaler definition it looks like this:
```yaml
apiVersion: custompodautoscaler.com/v1
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

For a `Deployment`:
```yaml
scaleTargetRef:
  apiVersion: apps/v1
  kind: Deployment
  name: hello-kubernetes
```
For a `ReplicaSet`:
```yaml
scaleTargetRef:
  apiVersion: apps/v1
  kind: ReplicaSet
  name: hello-kubernetes
```
For a `StatefulSet`:
```yaml
scaleTargetRef:
  apiVersion: apps/v1
  kind: StatefulSet
  name: hello-kubernetes
```
For a `ReplicationController`:
```yaml
scaleTargetRef:
  apiVersion: apps/v1
  kind: ReplicationController
  name: hello-kubernetes
```
