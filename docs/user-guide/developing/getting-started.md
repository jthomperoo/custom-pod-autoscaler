Developing a Custom Pod Autoscaler is designed to be an easy and flexible process, it can be done in any language of your preference and can use a wide variety of Docker images. For this guide we will build a simple Python based autoscaler, but you can take the principles outlined here and implement your autoscaler in any language, [see the examples for some other language implementations](https://github.com/jthomperoo/custom-pod-autoscaler/tree/master/example).

# Overview of guide

In this guide we will create a Python based Kubernetes autoscaler. The autoscaler will work by scaling the resource based on a label on the resouce `numPods`, will scale to the value provided in the label. This is not practical, but it should hopefully give a grounding and understanding of how to create a custom autoscaler.

# Set up the development environment

Dependencies required to follow this guide:

* [Python 3](https://www.python.org/downloads/)
* [Docker](https://docs.docker.com/install/)
* [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) or another Kubernetes cluster
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

# Create the project

Create a new directory for the project `python-custom-autoscaler` and begin working out the project directory.
```
mkdir python-custom-autoscaler
```

# Write the Custom Pod Autoscaler configuration

We will set up the Custom Pod Autoscaler configuration for this autoscaler, which will define which scripts to run, how to call them and a timeout for the script.  
Create a new file `config.yaml`:
```yaml
evaluate:
  type: "shell"
  timeout: 2500
  shell: "python /evaluate.py"
metric:
  type: "shell"
  timeout: 2500
  shell: "python /metric.py"
runMode: "per-resource"
```
This configuration file specifies the two scripts we are adding, the metric gatherer and the evaluator - defining that they should be called through a shell command, and they should timeout after `2500` milliseconds (2.5 seconds).  
The `runMode` is also specified, we have chosen `per-resource` - this means that the metric gathering will only run once for the resource, and the metric script will be provided the resource information. An alternative option is `per-pod`, which would run the metric gathering script for every pod the resource has, with the script being provided with individual pod information.

# Write the metric gatherer

We will now create the metric gathering part of the autoscaler, this part will simply read in the resource description provided and extract the `numPods` label value from it before outputting it back for the evaluator to make a decision with.  

Create a new file `metric.py`:
```python
import os
import json
import sys

def main():
    # Parse resource JSON into a dict
    resource = json.loads(sys.stdin.read())
    metric(resource)

def metric(resource):
    # Get metadata from resource information provided
    metadata = resource["metadata"]
    # Get labels from provided metdata
    labels = metadata["labels"]

    if "numPods" in labels:
        # If numPods label exists, output the value of the numPods 
        # label back to the autoscaler
        sys.stdout.write(labels["numPods"])
    else:
        # If no label numPods, output an error and fail the metric gathering
        sys.stderr.write("No 'numPods' label on resource being managed")
        exit(1)

if __name__ == "__main__":
    main()
```

The metric gathering stage gets relevant information piped into it from the autoscaler program; for this example we are running in `per-resource` mode - meaning the metric script is only called once for the resource being managed, and the resource information is piped into it. For example, if we are managing a deployment the autoscaler would provide a full JSON description of the deployment we are managing as the value piped in, e.g.
```json
{
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
        "creationTimestamp": "2019-12-22T17:16:53Z",
        "generation": 1,
        "labels": {
            "numPods": "1"
        },
        "name": "hello-kubernetes",
        "namespace": "default",
        "resourceVersion": "494588",
        "selfLink": "/apis/apps/v1/namespaces/default/deployments/hello-kubernetes",
        "uid": "63eeee05-a979-4573-b543-7d7dece9f431"
    },
    ...
}
```

The script we made simply parses this JSON and extracts out the `numPods` value. If `numPods` isn't provided, the script will error out and provide an error message which the autoscaler will pick up and log. If `numPods` is provided, the value will be output through standard out and read by the autoscaler, before being passed through to the evaluation stage.

# Write the evaluator

We will now create the evaluator, which will read in the metrics provided by the metric gathering stage and calculate the number of replicas based on these gathered metrics - in this case it will simply be the metric value provided by the previous step. If the value is not an integer, an error will be returned by the script.  

Create a new file `evaluate.py`:
```python
import json
import sys
import math

def main():
    # Parse metrics JSON into a dict
    metrics = json.loads(sys.stdin.read())
    evaluate(metrics)

def evaluate(metrics):
    try:
        value = int(metrics[0]["value"])

        # Build JSON dict with target_replicas
        evaluation = {}
        evaluation["target_replicas"] = value

        # Output JSON to stdout
        sys.stdout.write(json.dumps(evaluation))
    except ValueError as err:
        # If not an integer, output error
        sys.stderr.write(f"Invalid metric value: {err}")
        exit(1)

if __name__ == "__main__":
    main()

```

The JSON value piped into this step would look like this:
```json
{
  "deployment": "hello-kubernetes",
  "run_type": "scaler",
  "metrics": [
    {
      "resource": "hello-kubernetes",
      "value": "5"
    }
  ]
}
```
This is simply the metric value, in this case `5` from the previous step but wrapped in a JSON object, with additional information such as run type and deployment name.  

The JSON value output by this step would look like this:
```json
{
  "target_replicas": 5
}
```
The Custom Pod Autoscaler program expects the response to be in this JSON serialised form, with `target_replicas` defined as an integer.

# Write the Dockerfile

Our Dockerfile is going to be very simple, we will use the Python 3 docker image with the Custom Pod Autoscaler binary built into it.  
Create a new file `Dockerfile`:
```dockerfile
# Pull in Python build of CPA
FROM custompodautoscaler/python:latest

# Add config, evaluator and metric gathering Py scripts
ADD config.yaml evaluate.py metric.py /
```

This Dockerfile simply inserts in our two scripts and our configuration file. We have now finished creating our autoscaler, lets see how it works.

# Test the autoscaler

First we need to build the image for our new autoscaler, switch to the Minikube Docker registry as the target:
```
eval $(minikube docker-env)
```
Next build the image:
```
docker build -t python-custom-autoscaler .
```
Now we should deploy a resource to manage, create a deployment YAML file called `deployment.yaml`:
```yaml
```



# Modify the evaluation stage to scale to twice the label value

# Conclusion

Congratulations! You have now successfully created a custom Kubernetes autoscaler, for further information as to configuration options see the [configuration reference](reference/configuration.md), for more samples, check out the [examples for some other language implementations](https://github.com/jthomperoo/custom-pod-autoscaler/tree/master/example).
