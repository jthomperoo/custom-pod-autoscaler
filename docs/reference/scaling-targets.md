# Scaling targets

Custom Pod Autoscalers can target any resource that the Horizontal Pod Autoscaler can target, for example:

* Deployments
* ReplicaSets
* StatefulSets
* ReplicationControllers

Beyond this the CPA can target any other resource that your cluster supports as long as the resource implements the
scale subresource API. For example this means support for:

* Argo Rollouts
* Other third party resources.

## Scale target reference

To tell a Custom Pod Autoscaler which resource to target, provide a `scaleTargetRef` - a description of the resource to
target. Within a Custom Pod Autoscaler definition it looks like this:

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

For an Argo `Rollout`:

> Note: Argo Rollouts need to have a specialised Role, provide the `roleRequiresArgoRollouts: true` option to make sure
> the required role is provisioned.
> This feature is only available when using the [Custom Pod Autoscaler
Operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator) `v1.2.0` and above.

```yaml
scaleTargetRef:
  apiVersion: argoproj.io/v1alpha1
  kind: Rollout
  name: rollouts-demo
```
