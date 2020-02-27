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
import math

# JSON piped into this script example:
# {
#   "resource": "flask-metric",
#   "runType": "api",
#   "metrics": [
#     {
#       "resource": "flask-metric-869879868f-jgbg4",
#       "value": "{\"value\": 0, \"available\": 5, \"min\": 0, \"max\": 5}"
#     }
#   ]
# }

def main():
    # Parse JSON into a dict
    metrics = json.loads(sys.stdin.read())
    evaluate(metrics)

def evaluate(metrics):
    # Count total available
    total_available = 0
    for metric in metrics["metrics"]:
        json_value = json.loads(metric["value"])
        available = json_value["available"]
        total_available += int(available)

    # Get current replica count
    target_replica_count = len(metrics["metrics"])

    # Decrease target replicas if more than 5 available
    if total_available > 5:
        target_replica_count -= 1
    
    # Increase target replicas if none available
    if total_available <= 0:
        target_replica_count += 1

    # Build JSON dict with targetReplicas
    evaluation = {}
    evaluation["targetReplicas"] = target_replica_count

    # Output JSON to stdout
    sys.stdout.write(json.dumps(evaluation))

if __name__ == "__main__":
    main()
