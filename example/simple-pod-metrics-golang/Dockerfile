# Copyright 2020 The Custom Pod Autoscaler Authors.
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

FROM golang:1.13 AS build-env
ENV GO111MODULE=on
WORKDIR /build
# Copy in source and build binary
COPY . /build
RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -o simple-pod-metrics-golang

# Pull in Python build of CPA
FROM custompodautoscaler/alpine:latest
# Install dependencies for debugging
RUN apk add jq curl
# Set up aliases for easy debugging
RUN echo 'alias metrics="curl http://localhost:5000/api/v1/metrics | jq ."' >> ~/.profile
RUN echo 'alias evaluation="curl http://localhost:5000/api/v1/evaluation | jq ."' >> ~/.profile
# Add config
ADD config.yaml /
# Add built binary
COPY --from=build-env /build/simple-pod-metrics-golang /simple-pod-metrics-golang