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

import json
import sys

# JSON piped into this script example:
# [
#     {
#         "resource": "hello-kubernetes",
#         "value": "{\"up\": 3,\"down\": 2}"
#     },
# ]

def main():
    # Parse metrics in JSON
    metrics = json.loads(sys.stdin.read())
    evaluate(metrics)

def evaluate(metrics):
    # Only expect 1 metric provided
    if len(metrics) != 1:
        sys.stderr.write("Expected 1 metric")
        exit(1)

    # Get thumbs up and thumbs down values
    tweet_ratio_json = json.loads(metrics[0]["value"])
    up = int(tweet_ratio_json["up"])
    down = int(tweet_ratio_json["down"])

    # Calculate number of replicas
    replicas = up - down

    # Output target number of replicas to stdout
    evaluation = {}
    evaluation["target_replicas"] = replicas
    sys.stdout.write(json.dumps(evaluation))

if __name__ == "__main__":
    main()
