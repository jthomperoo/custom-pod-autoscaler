# Copyright 2021 The Custom Pod Autoscaler Authors.
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

FROM gradle:jdk11 AS build
RUN apt-get update
RUN apt-get install zip -y
COPY --chown=gradle:gradle . /home/gradle/src
WORKDIR /home/gradle/src
RUN gradle build --no-daemon
RUN cd /home/gradle/src/build/distributions/ && unzip /home/gradle/src/build/distributions/simple-pod-metrics-java.zip

# Pull in OpenJDK 11 build of CPA
FROM custompodautoscaler/openjdk-11:latest
# Install dependencies for debugging
RUN apt-get update
RUN apt-get install jq curl -y
# Set up aliases for easy debugging
RUN echo 'alias metrics="curl -X GET http://localhost:5000/api/v1/metrics | jq ."' >> ~/.bashrc
RUN echo 'alias evaluation="curl -X POST http://localhost:5000/api/v1/evaluation | jq ."' >> ~/.bashrc
# Add configuration file
ADD config.yaml /
# Add jar executable
COPY --from=build /home/gradle/src/build/distributions/simple-pod-metrics-java /app/simple-pod-metrics-java
