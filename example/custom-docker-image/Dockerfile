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

# Ubuntu base image
FROM ubuntu:latest

WORKDIR /app

# Install wget
RUN apt-get update && apt-get install wget -y

# Install CPA
RUN wget                                                                                                      \
    -qO-                                                                                                      \
    https://github.com/jthomperoo/custom-pod-autoscaler/releases/download/v2.1.0/custom-pod-autoscaler.tar.gz \
    | tar xvz                                                                                                 \
    && mv dist/* /app/

# Remove unused dist directory
RUN rm -r dist/

# Set up config
ENV configPath /app/config.yaml
COPY config.yaml ./

CMD [ "/app/custom-pod-autoscaler" ]
