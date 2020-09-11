# Custom Docker Image

This is an autoscaler that uses a custom Docker image that isn't provided by this repository (e.g. python, alpine).
This example uses the Ubuntu base image.

## Overview

The `Dockerfile` file defines how the image is built, using the `ubuntu` base image and downloading the Custom Pod
Autoscaler binary; setting it up to be executed on image run.

The autoscaler itself is simply a copy of the `python-custom-autoscaler` autoscaler from the
[Getting Started Guide](https://custom-pod-autoscaler.readthedocs.io/en/stable/user-guide/getting-started) with the
[full example autoscaler code here](../python-custom-autoscaler).

## Build

Build the image using:

```
docker build -t custom-docker-image .
```
