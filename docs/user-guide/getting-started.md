# Getting started
Developing a Custom Pod Autoscaler is designed to be an easy and flexible process, it can be done
in any language of your preference and can use a wide variety of Docker images. For this guide
we will build a simple Python based autoscaler, but you can take the principles outlined here
and implement your autoscaler in any language,
[see the examples for some other language implementations](https://github.com/jthomperoo/custom-pod-autoscaler/tree/master/example).

In this guide we will create a Python based Kubernetes autoscaler. The autoscaler will work by
scaling the resource based on a label on the resouce `numPods`, will scale to the value provided
in the label. This is not practical, but it should hopefully give a grounding and understanding
of how to create a custom autoscaler.
The finished and full code
[can be found in this example `python-custom-autoscaler`](https://github.com/jthomperoo/custom-pod-autoscaler/tree/master/example/python-custom-autoscaler).

# Set up the development environment

Dependencies required to follow this guide:

* [Python 3](https://www.python.org/downloads/)
* [Docker](https://docs.docker.com/install/)
* [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) or another Kubernetes cluster
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

# Create the project

Create a new directory for the project `python-custom-autoscaler` and begin working out the project
directory.
```
mkdir python-custom-autoscaler
```

# Write the Custom Pod Autoscaler configuration

We will set up the Custom Pod Autoscaler configuration for this autoscaler, which will define which
scripts to run, how to call them and a timeout for the script.
Create a new file `config.yaml`:
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
This configuration file specifies the two scripts we are adding, the metric gatherer and the evaluator
- defining that they should be called through a shell command, and they should timeout after `2500`
milliseconds (2.5 seconds).
The `runMode` is also specified, we have chosen `per-resource` - this means that the metric gathering
will only run once for the resource, and the metric script will be provided the resource information.
An alternative option is `per-pod`, which would run the metric gathering script for every pod the
resource has, with the script being provided with individual pod information.

# Write the metric gatherer

We will now create the metric gathering part of the autoscaler, this part will simply read in the
resource description provided and extract the `numPods` label value from it before outputting it
back for the evaluator to make a decision with.

Create a new file `metric.py`:
```python
import os
import json
import sys

def main():
    # Parse spec into a dict
    spec = json.loads(sys.stdin.read())
    metric(spec)

def metric(spec):
    # Get metadata from resource information provided
    metadata = spec["resource"]["metadata"]
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

The metric gathering stage gets relevant information piped into it from the autoscaler program; for
this example we are running in `per-resource` mode - meaning the metric script is only called once
for the resource being managed, and the resource information is piped into it. For example, if we
are managing a deployment the autoscaler would provide a full JSON description of the deployment
we are managing as the value piped in, e.g.
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

The script we made simply parses this JSON and extracts out the `numPods` value. If `numPods`
isn't provided, the script will error out and provide an error message which the autoscaler will
pick up and log. If `numPods` is provided, the value will be output through standard out and read
by the autoscaler, before being passed through to the evaluation stage.

# Write the evaluator

We will now create the evaluator, which will read in the metrics provided by the metric gathering
stage and calculate the number of replicas based on these gathered metrics - in this case it will
simply be the metric value provided by the previous step. If the value is not an integer, an error
will be returned by the script.

Create a new file `evaluate.py`:
```python
import json
import sys
import math

def main():
    # Parse provided spec into a dict
    spec = json.loads(sys.stdin.read())
    evaluate(spec)

def evaluate(spec):
    try:
        value = int(spec["metrics"][0]["value"])

        # Build JSON dict with targetReplicas
        evaluation = {}
        evaluation["targetReplicas"] = value

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
      "labels": {
        "numPods": "3"
      },
    },
    ...
  },
  "runType": "scaler"
}
```
This is simply the metric value, in this case `5` from the previous step but wrapped in a JSON
object, with additional information such as run type and deployment name.

The JSON value output by this step would look like this:

```json
{
  "targetReplicas": 5
}
```

The Custom Pod Autoscaler program expects the response to be in this JSON serialised form, with
`targetReplicas` defined as an integer.

# Write the Dockerfile

Our Dockerfile is going to be very simple, we will use the Python 3 docker image with the Custom Pod
Autoscaler binary built into it.
Create a new file `Dockerfile`:
```dockerfile
# Pull in Python build of CPA
FROM custompodautoscaler/python:latest

# Add config, evaluator and metric gathering Py scripts
ADD config.yaml evaluate.py metric.py /
```

This Dockerfile simply inserts in our two scripts and our configuration file. We have now finished
creating our autoscaler, lets see how it works.

# Test the autoscaler
First we should enable custom autoscalers on our cluster by installing the Custom Pod Autoscaler
Operator, for this guide we are using `v1.1.0`, but check out the latest version from the
[Custom Pod Autoscaler Operator
releases](https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases)
and see the [install
guide](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/master/INSTALL.md)
for the latest install information.

```bash
VERSION=v1.1.0
kubectl apply -f https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/cluster.yaml
```

This will do a cluster-wide install of `v1.1.0` of the Custom Pod Autoscaler Operator.

Now we should create a deployment for the autoscaler to manage, create a deployment YAML file
called `deployment.yaml`:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-kubernetes
  labels:
    numPods: "3"
spec:
  replicas: 1
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
This is a deployment we are going to manage with our custom autoscaler, it is initially set to have
`1` replica, but we have included `numPods: "3"` label in it, so if our autoscaler works it will
read this value and set the replica count to `3`.
We should run this deployment in the kubernetes cluster:
```
kubectl apply -f deployment.yaml
```
Using `kubectl get deployments` we should see our new deployment up and running with a single pod.

Now we need to build the image for our new autoscaler, switch to the Minikube Docker registry
as the target:
```
eval $(minikube docker-env)
```
Next build the image:
```
docker build -t python-custom-autoscaler .
```
Lets create the YAML for our autoscaler, create a new file `cpa.yaml`:
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
This YAML outlines the name of our autoscaler instance, the image/pod definition for our autoscaler
using `template`, the resource we are targeting to manage using `scaleTargetRef` and some basic
extra configuration with `interval` set to `10000` milliseconds, meaning the autoscaler will run
every 10 seconds.

We should now deploy our autoscaler to our cluster:
```
kubectl apply -f cpa.yaml
```

If we do `kubectl get pods` we should see a new pod has been created that is running our autoscaler.
Using `kubectl logs <POD NAME HERE> --follow` we should be able to see if any errors have occurred
running the autoscaler, alongside information as to how the autoscaler is deciding to scale.

After the autoscaler has scaled, it should increase the number of replicas for our managed
deployment to `3` - specified by the `numPods` label, check this with `kubectl get pods`.
Try updating the `deployment.yaml` label value and deploying it again, see how the replica
count changes.

# Modify the evaluation stage to scale to twice the label value

Now we have a working autoscaler, let's modify our evaluation decision making to set the replica
count to double the `numPods` label value.
First let's remove the autoscaler that is currently running by doing either:
```
kubectl delete cpa python-custom-autoscaler
```
or
```
kubectl delete -f cpa.yaml
```

Next let's modify the `evaluation.py` script to be like this:
```python
import json
import sys
import math

def main():
    # Parse provided spec into a dict
    spec = json.loads(sys.stdin.read())
    evaluate(spec)

def evaluate(spec):
    try:
        value = int(spec["metrics"][0]["value"])

        # Build JSON dict with targetReplicas
        evaluation = {}
        evaluation["targetReplicas"] = value * 2

        # Output JSON to stdout
        sys.stdout.write(json.dumps(evaluation))
    except ValueError as err:
        # If not an integer, output error
        sys.stderr.write(f"Invalid metric value: {err}")
        exit(1)

if __name__ == "__main__":
    main()

```

Rebuild the image as [described in the step above](#test-the-autoscaler).
Deploy the autoscaler again as [described in the step above](#test-the-autoscaler).
View the autoscaler logs using `kubectl get deployments` and `kubectl logs <POD_NAME_HERE> --follow`,
once an autoscale has occurred, check the number of pods for the managed resource using
`kubectl get pods`, it should have doubled.

# Clean up
Run these commands to remove any resouces created during this guide:

```bash
kubectl delete -f deployment.yaml
```

Removes our managed deployment.

```bash
kubectl delete -f cpa.yaml
```

Removes our custom autoscaler.

```bash
VERSION=v1.1.0
kubectl delete -f https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/cluster.yaml
```

Removes the custom autoscaler operator.

# Conclusion

Congratulations! You have now successfully created a custom Kubernetes autoscaler, for further
information as to configuration options see the
[configuration reference](../../../reference/configuration), for more samples, check out the
[examples for some other language implementations](https://github.com/jthomperoo/custom-pod-autoscaler/tree/master/example).
