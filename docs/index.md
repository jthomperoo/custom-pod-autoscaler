[![Build](https://github.com/jthomperoo/custom-pod-autoscaler/workflows/main/badge.svg)](https://github.com/jthomperoo/custom-pod-autoscaler/actions)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jthomperoo/custom-pod-autoscaler)
[![Go Report Card](https://goreportcard.com/badge/github.com/jthomperoo/custom-pod-autoscaler)](https://goreportcard.com/report/github.com/jthomperoo/custom-pod-autoscaler)
[![Documentation Status](https://readthedocs.org/projects/custom-pod-autoscaler/badge/?version=latest)](https://custom-pod-autoscaler.readthedocs.io/en/latest)
[![License](https://img.shields.io/:license-apache-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

# Custom Pod Autoscaler Framework

Visit the GitHub repository at <https://github.com/jthomperoo/custom-pod-autoscaler> to see examples, raise issues, and
to contribute to the project.

## What is the Custom Pod Autoscaler Framework?

Custom Pod Autoscalers (CPAs) are custom Kubernetes autoscalers. This project is part of a framework that lets you
quickly and easily build your own CPAs without having to deal with complex Kubernetes interactions using the tools and
language of your choice.

## Features

- Supports any language, environment and framework; the only requirement is it must be startable by a shell command
or HTTP request.
- Supports all configuration options of the Horizontal Pod Autoscaler (downscale stabilisation, sync period etc.)
- Allows fast and easy prototyping and development.
- Abstracts away all complicated Kubernetes API interactions.
- Exposes a HTTP REST API for integration with wider systems/manual intervention.
- Can write autoscalers with limited Kubernetes API or lifecycle knowledge.
- Configuration at build time or deploy time.
- Allows scaling to and from zero.
- Can be configured without master node access, can be configured on managed providers such as EKS or GKE.
- Supports Kubernetes metrics that the Horizontal Pod Autoscaler uses, can be configured using a similar syntax and
used in custom scaling decisions.
- Supports [Argo Rollouts](https://argoproj.github.io/argo-rollouts/).

## Why would I use it?

Kubernetes provides the Horizontal Pod Autoscaler, which allows automatic scaling of the number of replicas in a
resource (Deployment, ReplicationController, ReplicaSet, StatefulSet) based on metrics that you feed it. Mostly the
metrics used are CPU/memory load, which is sufficient for most applications. You can specify custom metrics to feed
into it through the metrics API also.

The limitation in the Horizontal Pod Autoscaler is that it has a [hard-coded algorithm for assessing these
metrics](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#algorithm-details):
```
desiredReplicas = ceil[currentReplicas * ( currentMetricValue / desiredMetricValue )]
```
If you need more flexibility in your scaling, beyond this algorithm, Horizontal Pod Autoscaler doesn't meet your needs,
you need to write your own scaling logic.

## What application would need custom scaling?

Taking an example from [Google Cloud's tutorial for hosting game servers on
Kubernetes](https://cloud.google.com/solutions/gaming/running-dedicated-game-servers-in-kubernetes-engine), there is a
section discussing autoscaling:
> The autoscaler can currently only scale the instance group based on CPU usage, which can be a misleading indicator of
> DGS load. Many DGSs are designed to consume idle cycles in an effort to optimize the game's simulation.

> As a result, many game developers implement a custom scaling manager process that is DGS aware to deal with the
> specific requirements of this type of workload.

The crux of the issue here is that for game servers, it doesn't make sense to scale on CPU load or memory usage, and
even if you implemented custom metrics the scaling algorithm wouldn't scale with these in a sensible way. The game
servers should scale on number of players on the servers, or number of players looking to join a server - trying to
ensure there are always positions available.

## How does it work?

The Custom Pod Autoscaler Framework provides a program that abstracts away complex Kubernetes interactions and handles
interacting with custom user logic you can provide to determine how the autoscaler should operate.

When developing a Custom Pod Autoscaler you define logic for two stages:

* Metric gathering - collecting or generating metrics; can be calling metrics APIs, running calculations locally,
making HTTP requests.
* Evaluating metrics - taking these gathered metrics and using them to decide how many replicas a resource should have.

These two pieces of logic are all the custom logic required to build a Custom Pod Autoscaler, the program will
handle all Kubernetes API interactions for scaling/retrieving resources.

See the [examples](https://github.com/jthomperoo/custom-pod-autoscaler/tree/master/example) or the [getting started
guide](user-guide/getting-started) for more information.
