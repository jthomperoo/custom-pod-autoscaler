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

import json
import sys
import requests

def main():
    # Parse JSON into a dict
    spec = json.loads(sys.stdin.read())
    metric(spec)

def metric(spec):
    # Get Pod IP
    status = spec["resource"]["status"]
    ip = status["podIP"]
    try:
        # Make request to Pod metric endpoint
        # (see ../flask-metrics/ folder for simple flask app exposing this endpoint)
        response = requests.get(f"http://{ip}:5000/metric")
        # Output whatever metrics are gathered to stdout
        sys.stdout.write(response.text)
    except HTTPError as http_err:
        # If an error occurs, output error to stderr and exit with status 1
        sys.stderr.write(f"HTTP error occurred: {http_err}")
        exit(1)
    except Exception as err:
        # If an error occurs, output error to stderr and exit with status 1
        sys.stderr.write(f"Other error occurred: {err}")
        exit(1)

if __name__ == "__main__":
    main()
