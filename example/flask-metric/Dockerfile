# Copyright 2025 The Custom Pod Autoscaler Authors.
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

FROM python

# Install dependencies for debugging
RUN apt-get update
RUN apt-get install jq curl -y
# Set up aliases for easy debugging
RUN echo 'alias metric="curl http://localhost:5000/metric | jq ."' >> ~/.bashrc
RUN echo 'alias increment="curl -X POST http://localhost:5000/increment | jq ."' >> ~/.bashrc
RUN echo 'alias decrement="curl -X POST http://localhost:5000/decrement | jq ."' >> ~/.bashrc
# We copy just the requirements.txt first to leverage Docker cache
COPY ./requirements.txt /app/requirements.txt
WORKDIR /app
# Install dependencies
RUN pip install -r requirements.txt
# Copy in source files
COPY . /app
ENTRYPOINT [ "python" ]
CMD [ "api.py" ]
