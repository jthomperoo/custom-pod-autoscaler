# Kubernetes Metrics

The Custom Pod Autoscaler supports automatically gathering useful metrics from Kubernetes for your autoscaler, which
is then fed into the metric gathering stage for your autoscaler to process and filter.

The pipeline looks like this:

1. Gather K8s resource information (Pod/Deployment/StatefulSet etc.).
2. Query the metrics server if the metrics server is available using the resource information retrieved.
3. If the `requireKubernetesMetrics` flag is set to `false` and it fails to get the metrics, continue as normal. If it
is set to `true` then fail at this point with a useful error.
4. Combine the K8s resource information with any gathered metrics information, serialising into JSON.
5. Call the metrics stage, providing the JSON generated in the previous step.
6. Get the results of the metrics stage, continue as normal.

The Custom Pod Autoscaler allows the autoscaler to define in its configuration any Kubernetes metrics it needs, for
example to get CPU details this configuration can be used:

```yaml
kubernetesMetricSpecs:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
```

Once this data has been fetched by the Custom Pod Autoscaler, it is then exposed to the metric gathering stage in
serialized JSON form, for example:

```json
{
  "resource": {...},
  "runType": "scaler",
  "kubernetesMetrics": [
    {
      "current_replicas": 1,
      "spec": {
        "type": "Resource",
        "resource": {
          "name": "cpu",
          "target": {
            "type": "Utilization"
          }
        }
      },
      "resource": {
        "pod_metrics_info": {
          "flask-metric-697794dd85-bsttm": {
            "Timestamp": "2021-04-05T18:10:10Z",
            "Window": 30000000000,
            "Value": 4
          }
        },
        "requests": {
          "flask-metric-697794dd85-bsttm": 200
        },
        "ready_pod_count": 1,
        "ignored_pods": {},
        "missing_pods": {},
        "total_pods": 1,
        "timestamp": "2021-04-05T18:10:10Z"
      }
    }
  ]
}
```

Visit the [configuration reference](../reference/configuration.md#kubernetesmetricspecs) for full details and the
[`k8s-metrics-cpu` example](https://github.com/jthomperoo/custom-pod-autoscaler/tree/master/example/k8s-metrics-cpu)
for sample code.
