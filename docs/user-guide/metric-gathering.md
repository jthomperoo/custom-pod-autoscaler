# Metric Gathering

The metric gathering stage is the first stage of the autoscaler. This stage is responsible for gathering all
information that the scaling decision will be based on. The autoscaler base program passes in information based on the
`runMode`, and the metric program is expected to output the metrics it gathers and to report any errors.

# Run modes

The `runMode` specifies how often the metric gathering is run, and the information supplied to the metric gatherer. How
the information is provided depends on the method used, e.g. if it is a shell command method the information will be
provided through standard in and piped to the script, see the [methods section for more information](../methods).

## per-pod

This mode runs the metric gatherer against every pod in the resource being managed.

Each individual pod's information is passed into the metric gatherer each time it is run against  each pod in JSON. An
example of the pod JSON passed to the metric gatherer:

```json
{
  "resource": {
      "apiVersion": "v1",
      "kind": "Pod",
      "metadata": {
          "creationTimestamp": "2019-11-14T01:52:21Z",
          "generateName": "flask-metric-869879868f-",
          "labels": {
              "app": "flask-metric",
              "pod-template-hash": "869879868f"
          },
          "name": "flask-metric-869879868f-2cslm",
          "namespace": "default",
          "ownerReferences": [
              {
                  "apiVersion": "apps/v1",
                  "blockOwnerDeletion": true,
                  "controller": true,
                  "kind": "ReplicaSet",
                  "name": "flask-metric-869879868f",
                  "uid": "2b028109-4793-4409-bd9c-a44d74da2fbc"
              }
          ],
          "resourceVersion": "208999",
          "selfLink": "/api/v1/namespaces/default/pods/flask-metric-869879868f-2cslm",
          "uid": "5a0ab9a6-dccc-497d-8d41-11f0408740b3"
      },
      ...
  },
  "runType": "scaler"
}
```

See the [Kubernetes Pod definition for full
description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#pod-v1-core).

## per-resource

This mode runs the metric gatherer only once against the resource being managed.

The resource information is passed into the metric gatherer in JSON, for example if managing a deployment the
deployment information will be passed in.

An example of deployment JSON passed to the metric gatherer:

```json
{
  "resource": {
    "kind": "Deployment",
    "apiVersion": "apps/v1",
    "metadata": {
      "name": "hello-kubernetes",
      "namespace": "default",
      "labels": {
        "numPods": "3"
      },
    },
    ...
  },
  "runType": "scaler"
}
```

See the [Kubernetes Deployment definition for full
description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#deployment-v1-apps).

# Kubernetes Metrics

The Custom Pod Autoscaler supports automatically querying the Kubernetes Metrics Server to gather metrics (e.g. CPU,
memory, custom metrics). These Kubernetes metrics are then serialized and included in the JSON sent to the metric
gathering stage.

For more details visit the [Kubernetes Metrics guide in this wiki](./kubernetes-metrics.md).

# Information output by the metric gatherer

The metric gatherer should gather/calculate the metrics it needs, and simply needs to return these
metrics back to the autoscaler, if an error occurs the metric gatherer should report this error,
alongside surfacing any error information.
How the metric gatherer returns information/errors is dependent on the method used for the metric
gatherer, for example if it is a shell command method, information is returned by writing to
standard out, and an error is signified by a non-zero exit code - with further error information
stored to standard error and standard out, see the [methods section for more information](../methods).
