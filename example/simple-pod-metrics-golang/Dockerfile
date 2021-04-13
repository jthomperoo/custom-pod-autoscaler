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

# Pull in Alpine build of CPA
FROM custompodautoscaler/alpine:latest
# Install dependencies for debugging
RUN apk add jq curl
# Set up aliases for easy debugging
RUN echo 'alias metrics="curl -X GET http://localhost:5000/api/v1/metrics | jq ."' >> ~/.profile
RUN echo 'alias evaluation="curl -X POST http://localhost:5000/api/v1/evaluation | jq ."' >> ~/.profile
# Add config and executable
ADD config.yaml dist/simple-pod-metrics-golang /
