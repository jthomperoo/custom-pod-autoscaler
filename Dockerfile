# Copyright 2022 The Custom Pod Autoscaler Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Python 3.6 build
FROM python:3.8-slim-buster AS python-3-8
WORKDIR /app
COPY dist /app/
CMD [ "/app/custom-pod-autoscaler" ]

# Python 3.7 build
FROM python:3.7-slim-buster AS python-3-7
WORKDIR /app
COPY dist /app/
CMD [ "/app/custom-pod-autoscaler" ]

# Python 3.6 build
FROM python:3.6-slim-buster AS python-3-6
WORKDIR /app
COPY dist /app/
CMD [ "/app/custom-pod-autoscaler" ]

# Alpine build
FROM alpine:3.10 AS alpine
WORKDIR /app
COPY dist /app/
CMD [ "/app/custom-pod-autoscaler" ]

# OpenJDK build
FROM openjdk:11-jre-slim-buster AS openjdk-11
WORKDIR /app
COPY dist /app/
CMD [ "/app/custom-pod-autoscaler" ]
