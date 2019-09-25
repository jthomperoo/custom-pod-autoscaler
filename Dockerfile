# Copyright 2019 The Custom Pod Autoscaler Authors.
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

# Global environment variables
ARG host=0.0.0.0 
ARG port=5000
ARG cmd="/app/custom-pod-autoscaler"

# Python build
FROM python:3.6-slim AS python
ARG host
ARG port
ARG cmd
ENV HOST=$host PORT=$port CMD=$cmd
WORKDIR /app
COPY dist /app/
CMD $CMD

# Alpine build
FROM alpine:3.10 AS alpine
ARG host
ARG port
ARG cmd
ENV HOST=$host PORT=$port CMD=$cmd
WORKDIR /app
COPY dist /app/
CMD $CMD