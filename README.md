[![Build](https://github.com/jthomperoo/custom-pod-autoscaler/workflows/main/badge.svg)](https://github.com/jthomperoo/custom-pod-autoscaler/actions)
[![codecov](https://codecov.io/gh/jthomperoo/custom-pod-autoscaler/branch/master/graph/badge.svg)](https://codecov.io/gh/jthomperoo/custom-pod-autoscaler)
[![GoDoc](https://godoc.org/github.com/jthomperoo/custom-pod-autoscaler?status.svg)](https://godoc.org/github.com/jthomperoo/custom-pod-autoscaler)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jthomperoo/custom-pod-autoscaler)
[![Go Report Card](https://goreportcard.com/badge/github.com/jthomperoo/custom-pod-autoscaler)](https://goreportcard.com/report/github.com/jthomperoo/custom-pod-autoscaler)
[![Documentation Status](https://readthedocs.org/projects/custom-pod-autoscaler/badge/?version=latest)](https://custom-pod-autoscaler.readthedocs.io/en/latest)
[![License](http://img.shields.io/:license-apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
# Custom Pod Autoscaler

This is the Custom Pod Autoscaler (CPA) code and base images.  

The CPA is part of the [Custom Pod Autoscaler Framework](https://custom-pod-autoscaler.readthedocs.io/en/latest).  

## What is it?

A Custom Pod Autoscaler is a Kubernetes autoscaler that is customised and user created. Custom Pod Autoscalers are designed to be similar to the Kubernetes Horizontal Pod Autoscaler. The Custom Pod Autoscaler framework allows easier and faster development of Kubernetes autoscalers.  
A Custom Pod Autoscaler can be created by using this project, extending the Docker base images provided and inserting your own logic; see the [examples for more information](https://github.com/jthomperoo/custom-pod-autoscaler/tree/master/example).  

## How does it work?
A Custom Pod Autoscaler has a base program (defined in this repository) that handles interacting with user logic, for example by using shell commands and piping data into them.  
When developing a Custom Pod Autoscaler you define logic for two stages:

* Metric gathering - collecting or generating metrics; can be calling metrics APIs, running calculations locally, making HTTP requests.
* Evaluating metrics - taking these gathered metrics and using them to decide how many replicas a resource should have.

These two pieces of logic are all the custom logic required to build a Custom Pod Autoscaler, the base program will handle all Kubernetes API interactions for scaling/retrieving resources.

## Getting started

Check out [this getting started guide for a quick start for developers](https://custom-pod-autoscaler.readthedocs.io/en/latest/user-guide/getting-started).

## More information

See the [wiki for more information, such as guides and references](https://custom-pod-autoscaler.readthedocs.io/en/latest/).

## Developing this project
### Environment
Developing this project requires these dependencies:

* [Go](https://golang.org/doc/install) >= `1.13`
* [Golint](https://github.com/golang/lint)
* [Docker](https://docs.docker.com/install/)

To view docs locally, requires:

* [mkdocs](https://www.mkdocs.org/)

### Commands

* `make` - builds the CPA binary.
* `make docker` - builds the CPA base images.
* `make lint` - lints the code.
* `make unittest` - runs the unit tests.
* `make vendor` - generates a vendor folder.
* `make doc` - hosts the documentation locally, at `127.0.0.1:8000`.