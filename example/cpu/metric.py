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

import json
import sys
import requests

# JSON piped into this script example:
# {
#     ...
#     "kubernetesMetrics": [
#         {
#             "current_replicas": 1,
#             "spec": {
#                 "type": "Resource",
#                 "resource": {
#                     "name": "cpu",
#                     "target": {
#                         "type": "Utilization"
#                     }
#                 }
#             },
#             "resource": {
#                 "pod_metrics_info": {
#                     "flask-metric-697794dd85-bsttm": {
#                         "Timestamp": "2021-04-05T18:10:10Z",
#                         "Window": 30000000000,
#                         "Value": 4
#                     }
#                 },
#                 "requests": {
#                     "flask-metric-697794dd85-bsttm": 200
#                 },
#                 "ready_pod_count": 1,
#                 "ignored_pods": {},
#                 "missing_pods": {},
#                 "total_pods": 1,
#                 "timestamp": "2021-04-05T18:10:10Z"
#             }
#         }
#     ]
#     ...
# }

def main():
    # Parse JSON into a dict
    spec = json.loads(sys.stdin.read())
    metric(spec)

def metric(spec):
    cpu_metrics = spec["kubernetesMetrics"][0]
    current_replicas = cpu_metrics["current_replicas"]
    resource = cpu_metrics["resource"]
    pod_metrics_info = resource["pod_metrics_info"]
    total_utilization = 0
    for _, pod_info in pod_metrics_info.items():
        total_utilization += pod_info["Value"]
    average_utilization = total_utilization / current_replicas
    sys.stdout.write(json.dumps(
        {
            "current_replicas": current_replicas,
            "average_utilization": average_utilization
        }
    ))

if __name__ == "__main__":
    main()
