[![Build](https://github.com/jthomperoo/custom-pod-autoscaler/workflows/main/badge.svg)](https://github.com/jthomperoo/custom-pod-autoscaler/actions)
[![codecov](https://codecov.io/gh/jthomperoo/custom-pod-autoscaler/branch/master/graph/badge.svg)](https://codecov.io/gh/jthomperoo/custom-pod-autoscaler)
[![GoDoc](https://godoc.org/github.com/jthomperoo/custom-pod-autoscaler?status.svg)](https://godoc.org/github.com/jthomperoo/custom-pod-autoscaler)
[![Go Report Card](https://goreportcard.com/badge/github.com/jthomperoo/custom-pod-autoscaler)](https://goreportcard.com/report/github.com/jthomperoo/custom-pod-autoscaler)
[![License](http://img.shields.io/:license-apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
# Custom Pod Autoscaler

This is the Custom Pod Autoscaler (CPA) code and base images.  

The CPA is part of the [Custom Pod Autoscaler Framework](https://github.com/jthomperoo/custom-pod-autoscaler/wiki/Custom-Pod-Autoscaler-Framework).  
## Use

The CPA can be used to create custom scaling logic for Kubernetes; similar to the Horizontal Pod Autoscaler, but allowing the use of user defined commands and scripts to manage metric gathering and evaluating how many replicas to scale to.

## Developing your own Custom Pod Autoscaler

Custom Pod Autoscalers are Docker images that are designed to run in a Kubernetes cluster, you can build your own either by extending one of the Docker images provided, or by taking the CPA binary and inserting it into your own Docker image.  

Docker images provided:

* `custompodautoscaler/python` - Image set up to run Python 3 scripts
* `custompodautoscaler/alpine` - Minimal alpine image.

CPA binaries are distributed in [GitHub Releases as an asset](https://github.com/jthomperoo/custom-pod-autoscaler/releases), inside `custom-pod-autoscaler.tar.gz`.

See the [example for more information](https://github.com/jthomperoo/custom-pod-autoscaler/tree/master/example).


## Installing a Custom Pod Autoscaler

The easiest way to install a CPA is using the [Custom Pod Autoscaler Operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator), follow the [installation guide for instructions for installing the operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/master/INSTALL.md).


## Developing this project
### Environment
Developing this project requires these dependencies:

* Go >= 1.13
* Golint
* Docker

### Commands

* `make` - builds the CPA binary.
* `make docker` - builds the CPA base images.
* `make lint` - lints the code.
* `make unittest` - runs the unit tests.
* `make vendor` - generates a vendor folder.