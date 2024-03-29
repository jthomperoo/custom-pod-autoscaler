# Configuration

Configuration is a key part of the Custom Pod Autoscaler, defining how to call user logic, alongside more fine tuned
configuration options that the end user might adjust; such as polling interval.

## How to provide configuration

### Configuration file in image

Example:

```yaml
evaluate:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/evaluate.py"
metric:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/metric.py"
runMode: "per-resource"
```

Configuration may be set by using a configuration file (default path `/config.yaml`), this can be baked into a Docker
image to provide standard configuration for the autoscaler - for example metric gathering and evaluation methods.

### Configuration passed as environment variables (defined in deployment YAML)

Example:

```yaml
  config:
    - name: interval
      value: "10000"
    - name: startTime
      value: "60000"
```

Configuration may be passed as environment variables; these can be customised at deploy time rather than baked in at
build time, so allow for more fine tuned customisation. The main way to define the environment variables should be
using the `config` YAML description, which allows for key-value pairs to define each configuration option and the value
it should have. All configuration options defined as key-value pairs in the `config` YAML are converted into environment
variables by the operator; this allows autoscalers to extend the configuration options and use this `config` YAML to
define configuration for the user logic.


> Note: Configuration set as an environment variable takes precedence over configuration set in a configuration file,
> this allows the configuration file to act possibly as a set of defaults that can be overridden at deploy time.

## Methods

Example:

```yaml
metric:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/metric.py"
```

A method defines a hook for calling user logic, with a timeout to handle user
logic that hangs.

### type

- `shell` = call the user logic through a shell command, with the relevant
  information piped through to the command. The user logic communicates back
  with the autoscaler through exit codes and standard error and out. A non zero
  exit code tells the autoscaler that the user logic has failed; and the
  autoscaler will read in standard error and log it for debug purposes. If no
  error occurs, the autoscaler may read in the standard out and use it, e.g. for
  metric gathering.
- `http` = make an HTTP request to retrieve the results, with the relevant
  information passed through via a parameter (either body or query). The user
  logic communicates back with the autoscaler through HTTP status codes and the
  response body. A status code that is defined as not successful (configured as
  part of the method) tells the autoscaler that the request has failed; while a
  successful response code indicates success.

Defines the type of the method.

### timeout

Defines how long the autoscaler should wait for the user logic to finish, if it exceeds this time it will assume the
operation has failed and provide a timeout error.

### shell

Defines a shell method, which is a simple string with the shell command to execute. Shell commands are configured with
entrypoints and commands to execute, with values piped in through standard in. See the [methods section for more
details](../user-guide/methods.md#shell)

### http

Defines an http method, which configures how an HTTP request is made. See the [methods section for more
details](../user-guide/methods.md#http)

## configPath

```yaml
config:
  - name: configPath
    value: "/config.yaml"
```

Default: `/config.yaml`

This defines the path to the configuration file. Should only be defined as an environment variable (through deployment
YAML), as defining the path to the configuration file inside the configuration file does not make sense.

## interval

Example:

```yaml
interval: 15000
```

Default value: `15000`

This defines in milliseconds how frequently the autoscaler should run. The autoscaler will run, then wait this many
milliseconds before running again.

## evaluate

Example:

```yaml
evaluate:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/evaluate.py"
```

No default value, required to be set.

This defines the evaluation logic that should be run, and how it should be triggered.
[This is a `method`, see methods section for full configuration options of a method](#methods).

## metric

Example:

```yaml
metric:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/metric.py"
```

No default value, required to be set.

This defines the metric logic that should be run, and how it should be triggered.
[This is a `method`, see methods section for full configuration options of a method](#methods).

## namespace

Example:

```yaml
namespace: "default"
```

Default value: `default`

Defines the namespace to look in for the resource being managed.

## minReplicas

Example:

```yaml
minReplicas: 1
```

Default value: `1`

Defines the minimum number of replicas allowed, resource won't be scaled below this value. If set to anything but `0`
scaling to zero is disabled, and a replica count of `0` will be treated as autoscaling disabled. If set to `0` scaling
to `0` is enabled.

## maxReplicas

Example:

```yaml
maxReplicas: 10
```

Default value: `10`

Defines the maximum number of replicas allowed, resource won't be scaled above this value.

## runMode

Example:

```yaml
runMode: "per-pod"
```

Default value: `per-pod`

- `per-pod` = runs metric gathering per pod, individually running the user logic for each pod in the resource being
managed, with the pod information provided to the user logic.
- `per-resource` = runs metric gathering per resource, running the user logic for only the resource being managed, with
the resource information provided to the user logic.

Defines how the autoscaler runs the metric gathering user logic, changing the values that are provided to the metric
gathering user logic and changing how frequently the metric gathering user logic is called.

## startTime

Example:

```yaml
startTime: 1
```

Default value: `1`

This defines in milliseconds a starting point for the scaler, with the scaler running as if it started at the time
provided. Allows specifying that the autoscaler must start on a multiple of the interval from the start time. For
example, a startTime of `60000` would result in the autoscaler starting at the next full minute. The default value will
start the autoscaler after a single millisecond, close to instantly.

> Note: The scaling will not actually start until after one interval has passed.

## logVerbosity

Example:

```yaml
logVerbosity: 0
```

Default value: `0`

This defines the verbosity of the logging, allowing for debugging errors/issues. Logging will occur for all values
ABOVE and including the verbosity level set.

Log levels:

* `0` - normal.
* `1` - verbose.
* `2` - more verbose around high level logic such as autoscaling/rest api.
* `3` - more verbose around lower level logic such as metric gathering and evaluation.

## downscaleStabilization

Example:

```yaml
downscaleStabilization: 200
```

Default value: `0`

This defines in seconds the length of the downscale stabilization window; based on [the Horizontal Pod Autoscaler
downscale
stabilization](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-cooldown-delay).
Downscale stabilization works by recording all evaluations over the window specified and picking out the maximum target
replicas from these evaluations. This results in a more smoothed downscaling and a cooldown, which can reduce the effect
of thrashing.

## apiConfig

Example:

```yaml
apiConfig:
  enabled: true
  useHTTPS: true
  port: 80
  host: "0.0.0.0"
  certFile: "cert.crt"
  keyFile: "key.key"
```

Default value:

```yaml
apiConfig:
  enabled: true
  useHTTPS: false
  port: 5000
  host: "0.0.0.0"
  certFile: ""
  keyFile: ""
```

This sets configuration options for the Custom Pod Autoscaler API, allowing enabling and disabling the API and
specifying how the API should be exposed.

### enabled

Boolean value to enable (`true`) or disable (`false`) the API

### useHTTPS

Boolean value to enable (`true`) or disable (`false`) HTTPS.

### port

Integer value defining the port to expose the API on.

### host

String value defining the host to expose the API on.

### certFile

String value defining the path to the [certificate](https://golang.org/pkg/net/http/#ListenAndServeTLS) to use for
HTTPS.

### keyFile

String value defining the path to the [private key](https://golang.org/pkg/net/http/#ListenAndServeTLS) to use for
HTTPS.

## preMetric

Example:

```yaml
preMetric:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/pre-metric.py"
```

No default value, if not set it is not executed.

This defines a pre-metric hook, and how it should be triggered.
The pre-metric hook is run before metric gathering occurs, it is provided with either the resource being managed or the
pods being managed (depending on the run mode) as JSON.

Example of JSON provided to this hook:

```json
{
  "resource": {
    "kind": "Deployment",
    "apiVersion": "apps/v1",
    "metadata": {
      "name": "hello-kubernetes",
      "namespace": "default",
    },
    ...
  },
  "runType": "scaler"
}
```

[This is a `method`](#methods) that is running as part of a [`hook`](../user-guide/hooks.md).

## postMetric

Example:

```yaml
postMetric:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/post-metric.py"
```

No default value, if not set it is not executed.

This defines a post-metric hook, and how it should be triggered.
The post-metric hook is run after metric gathering occurs, it is provided with either the resource being managed or the
pods being managed (depending on the run mode) alongside the metric gathering results as JSON.

Example of JSON provided to this hook:
```json
{
  "resource": {
    "kind": "Deployment",
    "apiVersion": "apps/v1",
    "metadata": {
      "name": "hello-kubernetes",
      "namespace": "default",
    },
    ...
  },
  "metrics": [
    {
      "resource": "hello-kubernetes",
      "value": "3"
    }
  ],
  "runType": "scaler"
}
```

[This is a `method`](#methods) that is running as part of a [`hook`](../user-guide/hooks.md).

## preEvaluate

Example:

```yaml
preEvaluate:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/pre-evaluate.py"
```

No default value, if not set it is not executed.

This defines a pre-evaluate hook, and how it should be triggered.
The pre-evaluate hook is run before evaluation occurs, it is provided with the full resource metrics as JSON.

Example of JSON provided to this hook:
```json
{
  "metrics": [
    {
      "resource": "hello-kubernetes",
      "value": "3"
    }
  ],
  "resource": {
    "kind": "Deployment",
    "apiVersion": "apps/v1",
    "metadata": {
      "name": "hello-kubernetes",
      "namespace": "default",
    },
    ...
  },
  "runType": "scaler"
}
```

[This is a `method`](#methods) that is running as part of a [`hook`](../user-guide/hooks.md).

## postEvaluate

Example:

```yaml
postEvaluate:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/post-evaluate.py"
```

No default value, if not set it is not executed.

This defines a post-evaluate hook, and how it should be triggered.
The post-evaluate hook is run after evaluation occurs, it is provided with the full resource metrics alongside the
evaluation that has been calculated as JSON.

Example of JSON provided to this hook:
```json
{
  "metrics": [
    {
      "resource": "hello-kubernetes",
      "value": "3"
    }
  ],
  "resource": {
    "kind": "Deployment",
    "apiVersion": "apps/v1",
    "metadata": {
      "name": "hello-kubernetes",
      "namespace": "default",
    },
    ...
  },
  "evaluation": {
    "targetReplicas": 3
  },
  "runType": "scaler"
}
```

[This is a `method`](#methods) that is running as part of a [`hook`](../user-guide/hooks.md).

## preScale

Example:

```yaml
preScale:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/pre-scale.py"
```

No default value, if not set it is not executed.

This defines a pre-scaling hook, and how it should be triggered.
This hook is run even if autoscaling is disabled for the resource (replicas set to `0`).
The pre-scale hook is run before a scaling decision is made, it is provided with min and max replicas, current replicas,
target replicas, and resource being scaled as JSON.

Example of JSON provided to this hook:

```json
{
  "evaluation": {
    "targetReplicas": 6
  },
  "resource": {
    "kind": "Deployment",
    "apiVersion": "apps/v1",
    "metadata": {
      "name": "hello-kubernetes",
      "namespace": "default",
    },
    ...
  },
  "scaleTargetRef": {
    "kind": "Deployment",
    "name": "hello-kubernetes",
    "apiVersion": "apps/v1"
  },
  "namespace": "default",
  "minReplicas": 1,
  "maxReplicas": 10,
  "targetReplicas": 0,
  "runType": "scaler"
}
```

[This is a `method`](#methods) that is running as part of a [`hook`](../user-guide/hooks.md).

## postScale

Example:

```yaml
postScale:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/post-scale.py"
```

No default value, if not set it is not executed.

This defines a post-scaling hook, and how it should be triggered.
This hook is only run if scaling is successful, and is not run if autoscaling is disabled for the resource (replicas
set to `0`).

The post-scale hook is run after a scaling decision is made and effected, it is provided with min and max replicas,
current replicas, target replicas, and resource being scaled as JSON.

Example of JSON provided to this hook:

```json
{
  "evaluation": {
    "targetReplicas": 6
  },
  "resource": {
    "kind": "Deployment",
    "apiVersion": "apps/v1",
    "metadata": {
      "name": "hello-kubernetes",
      "namespace": "default",
    },
    ...
  },
  "scaleTargetRef": {
    "kind": "Deployment",
    "name": "hello-kubernetes",
    "apiVersion": "apps/v1"
  },
  "namespace": "default",
  "minReplicas": 1,
  "maxReplicas": 10,
  "targetReplicas": 0,
  "runType": "scaler"
}
```

[This is a `method`](#methods) that is running as part of a [`hook`](../user-guide/hooks.md).

## kubernetesMetricSpecs

Example:

```yaml
kubernetesMetricSpecs:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
  - type: Resource
    resource:
      name: memory
      target:
        type: AverageValue
```

No default value, if not set the Kubernetes metrics server will not be queried.

This defines a list of the Kubernetes metrics to fetch from the Kubernetes metrics server. These metrics are then
serialized and then fed into the metric gathering stage of the Custom Pod Autoscaler pipeline.

This option supports any Kubernetes metrics that can be defined as part of a Kubernetes Horizontal Pod Autoscaler:

> Note: Since the Custom Pod Autoscaler leaves evaluation of the metrics to the user logic, there is no need to provide
> target values for defining metric specs, so the `averageValue`, `value`, and `averageUtilization` values can be
> left out

For more information about the Kubernetes metrics and how to define which Kubernetes metrics the Custom Pod Autoscaler
needs see the Kubernetes documentation here:

<https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/#autoscaling-on-multiple-metrics-and-custom-metrics>


The metrics will be serialized in the following form:

```json
...
"kubernetesMetrics": [
  {
    "current_replicas": <current replicas>,
    "spec": <the metric spec used to generate the metrics in JSON>,
    "resource": <if it is a resource metric, this will be populated, otherwise it will not be set>,
    "pods": <if it is a pods metric, this will be populated, otherwise it will not be set>,
    "object": <if it is a object metric, this will be populated, otherwise it will not be set>,
    "external": <if it is a external metric, this will be populated, otherwise it will not be set>
  }
]
...
```

Each type of metric has a different JSON structure which must be processed differently, since they return different
sets of data.

### resource

Example of a `resource` metric spec:

```yaml
- type: Resource
  resource:
    name: cpu
    target:
      type: Utilization
```

The `resource` metric will be serialized into the following form:

```json
"resource": {
  "pod_metrics_info": <list of pod metrics>,
  "requests": <map of pod names to their request value of the resource being queried (e.g. CPU requests)>,
  "ready_pod_count": <number of ready pods>,
  "ignored_pods": <a set of ignored pod names>,
  "missing_pods": <a set of missing pod names>,
  "total_pods": <number of ready pods>,
  "timestamp": <timestamp>
}
```

Example of the serialised JSON Kubernetes metrics provided to the metric gathering stage if metric specs are provided:

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

### pods

Example of a `pods` metric spec:

```yaml
- type: Pods
  pods:
    metric:
      name: packets-per-second
    target:
      type: AverageValue
```

The `pods` metric will be serialized into the following form:

```json
"pods": {
  "pod_metrics_info": <list of pod metrics>,
  "ready_pod_count": <number of ready pods>,
  "ignored_pods": <a set of ignored pod names>,
  "missing_pods": <a set of missing pod names>,
  "total_pods": <number of ready pods>,
  "timestamp": <timestamp>
}
```

Example of the serialised JSON Kubernetes metrics provided to the metric gathering stage if metric specs are provided:

```json
{
  "resource": {...},
  "runType": "scaler",
  "kubernetesMetrics": [
    {
      "current_replicas": 1,
      "spec": {
        "type": "Pods",
        "pods": {
          "metric": {
            "name": "packets-per-second",
          },
          "target": {
            "type": "AverageValue"
          }
        }
      },
      "resource": {
        "pod_metrics_info": {
          "flask-metric-697794dd85-bsttm": {
            "Timestamp": "2021-04-05T18:10:10Z",
            "Window": 30000000000,
            "Value": 1000
          }
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

### object

Example of an `object` metric spec:

```yaml
- type: Object
  object:
    metric:
      name: requests-per-second
    describedObject:
      apiVersion: networking.k8s.io/v1beta1
      kind: Ingress
      name: main-route
    target:
      type: Value
```

The `object` metric will be serialized into the following form:

```json
"object": {
  "current": <the current value of the metric (value or average value)>,
  "ready_pod_count": <number of ready pods>,
  "timestamp": <timestamp>
}
```

Example of the serialised JSON Kubernetes metrics provided to the metric gathering stage if metric specs are provided:

```json
{
  "resource": {...},
  "runType": "scaler",
  "kubernetesMetrics": [
    {
      "current_replicas": 1,
      "spec": {
        "type": "External",
        "metric": {
          "name": "queue_messages_ready",
          "selector": "queue=worker_tasks",
          "target": {
            "type": "AverageValue"
          }
        }
      },
      "external": {
        "current": {
          "average_value": 5
        },
        "ready_pod_count": 1,
        "timestamp": "2021-04-05T18:10:10Z"
      }
    }
  ]
}
```

### external

Example of an `external` metric spec:

```yaml
- type: External
  external:
    metric:
      name: queue_messages_ready
      selector: "queue=worker_tasks"
    target:
      type: AverageValue
```

The `external` metric will be serialized into the following form:

```json
"external": {
  "current": <the current value of the metric (value or average value)>,
  "ready_pod_count": <number of ready pods>,
  "timestamp": <timestamp>
}
```

Example of the serialised JSON Kubernetes metrics provided to the metric gathering stage if metric specs are provided:

```json
{
  "resource": {...},
  "runType": "scaler",
  "kubernetesMetrics": [
    {
      "current_replicas": 1,
      "spec": {
        "type": "External",
        "metric": {
          "name": "queue_messages_ready",
          "selector": "queue=worker_tasks",
          "target": {
            "type": "Value"
          }
        }
      },
      "external": {
        "current": {
          "value": 5
        },
        "ready_pod_count": 1,
        "timestamp": "2021-04-05T18:10:10Z"
      }
    }
  ]
}
```

## requireKubernetesMetrics

Example:

```yaml
requireKubernetesMetrics: true
```

Default value: `false`.

Flag that allows the Custom Pod Autoscaler to fail if it does not successfully query the Kubernetes metrics server.
If the flag is set to `true` then the Custom Pod Autoscaler will fail with an error before sending data to the metric
gathering stage if the metrics server is unavailable/an error occurs querying the metrics server. If the flag is set to
`false` then if the Custom Pod Autoscaler fails to query the Kubernetes metrics server it will continue, logging the
error but not exiting.
