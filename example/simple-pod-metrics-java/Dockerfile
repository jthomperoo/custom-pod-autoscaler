# Copyright 2023 The Custom Pod Autoscaler Authors.
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

FROM gradle:jdk21 AS build
RUN apt-get update
RUN apt-get install zip -y
COPY --chown=gradle:gradle . /home/gradle/src
WORKDIR /home/gradle/src
RUN gradle build --no-daemon
RUN cd /home/gradle/src/build/distributions/ && unzip /home/gradle/src/build/distributions/simple-pod-metrics-java.zip

# Ubuntu base image
FROM ubuntu:latest

WORKDIR /app

# Install wget
RUN apt-get update && apt-get install wget jq curl openjdk-21-jre -y

# Install CPA
RUN wget                                                                                                      \
    -qO-                                                                                                      \
    https://github.com/jthomperoo/custom-pod-autoscaler/releases/download/v2.12.0/custom-pod-autoscaler.tar.gz \
    | tar xvz                                                                                                 \
    && mv dist/* /app/

# Remove unused dist directory
RUN rm -r dist/

# Set up aliases for easy debugging
RUN echo 'alias metrics="curl -X GET http://localhost:5000/api/v1/metrics | jq ."' >> ~/.bashrc
RUN echo 'alias evaluation="curl -X POST http://localhost:5000/api/v1/evaluation | jq ."' >> ~/.bashrc

# Add configuration file
ADD config.yaml /
# Add jar executable
COPY --from=build /home/gradle/src/build/distributions/simple-pod-metrics-java /app/simple-pod-metrics-java

CMD [ "/app/custom-pod-autoscaler" ]
