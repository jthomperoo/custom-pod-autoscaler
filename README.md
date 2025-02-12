[![Build](https://github.com/jthomperoo/custom-pod-autoscaler/workflows/main/badge.svg)](https://github.com/jthomperoo/custom-pod-autoscaler/actions)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jthomperoo/custom-pod-autoscaler/v2)
[![Go Report
Card](https://goreportcard.com/badge/github.com/jthomperoo/custom-pod-autoscaler)](https://goreportcard.com/report/github.com/jthomperoo/custom-pod-autoscaler/v2)
[![Documentation
Status](https://readthedocs.org/projects/custom-pod-autoscaler/badge/?version=stable)](https://custom-pod-autoscaler.readthedocs.io/en/stable)
[![License](https://img.shields.io/:license-apache-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

<p>This project is supported by:</p>
<p>
  <a href="https://www.businesssimulations.com/">
    <img src="https://businesssimulations.com/img/logos/logo_nostrap_tp_410.webp">
  </a>
</p>

# Custom Pod Autoscaler

Custom Pod Autoscalers (CPAs) are custom Kubernetes autoscalers. This project is part of a framework that lets you
quickly and easily build your own CPAs without having to deal with complex Kubernetes interactions using the tools and
language of your choice.

## What is this project?

This project is part of the Custom Pod Autoscaler Framework (CPAF) which is a set of tools to help you easily build
your own CPAs. This project is the core of the CPAF, providing a program which runs inside your CPA to manage
Kubernetes interactions and custom user logic interactions.

A Custom Pod Autoscaler can be created by using this project, extending the Docker base images provided and inserting
your own logic; see the [examples for more
information](https://github.com/jthomperoo/custom-pod-autoscaler/tree/v2.12.0/example).

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

## How does it work?

This project is a program that abstracts away complex Kubernetes interactions and handles interacting with custom
user logic you can provide to determine how the autoscaler should operate.

When developing a Custom Pod Autoscaler you define logic for two stages:

* Metric gathering - collecting or generating metrics; can be calling metrics APIs, running calculations locally,
making HTTP requests.
* Evaluating metrics - taking these gathered metrics and using them to decide how many replicas a resource should have.

These two pieces of logic are all the custom logic required to build a Custom Pod Autoscaler, the program will
handle all Kubernetes API interactions for scaling/retrieving resources.

## Getting started

Check out [this getting started guide for a quick start for
developers](https://custom-pod-autoscaler.readthedocs.io/en/stable/user-guide/getting-started).

## More information

See the [wiki for more information, such as guides and
references](https://custom-pod-autoscaler.readthedocs.io/en/stable/).

### What other projects are in the Custom Pod Autoscaler Framework?

The [Custom Pod Autoscaler Operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator) is the other part
of the Custom Pod Autoscaler Framework, it is an operator that handles provisioning Kubernetes resources for your
CPA.

## Developing this project

See the [contribution guidelines](./CONTRIBUTING.md).
