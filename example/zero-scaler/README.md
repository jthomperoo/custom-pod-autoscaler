# Zero Scaler

This is a simple Python based autoscaler, designed to show that the CPAF
supports scaling to and from `0`.  
This is a modified version of the
[`python-custom-autoscaler`](../python-custom-autoscaler) which is talked
through in [the getting started
guide.](https://custom-pod-autoscaler.readthedocs.io/en/latest/user-guide/getting-started).

## Overview

This is a simple autoscaler, it will check the resource it is managing for the
label `numPods`, for example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-kubernetes
  labels:
    numPods: "3"
...
```

The autoscaler will read this `numPods` label and scale to its value, set the
`replicas` value to a different value than this `numPods` to check that scaling
works correctly, for example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-kubernetes
  labels:
    numPods: "3"
spec:
  replicas: 5
  selector:
    matchLabels:
      app: hello-kubernetes
  template:
    metadata:
      labels:
        app: hello-kubernetes
    spec:
      containers:
      - name: hello-kubernetes
        image: paulbouwer/hello-kubernetes:1.5
        ports:
        - containerPort: 8080
```

This would start with a replica count of `5`, before the autoscaler kicks in and
reads the `numPods` value of `3` and scales to `3` replicas.

## Zero Scaling

This autoscaler is able to scale to and from `0` because in `config.yaml` the
`minReplicas` is set to `0`.