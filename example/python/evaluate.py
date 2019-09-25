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
import requests
import math

def main():
    metric_json = json.loads(sys.stdin.read())
    evaluate(metric_json)

def evaluate(metrics):
    total_available = 0
    for metric in metrics:
        pod = metric["pod"]
        available = metric["available"]
        total_available += int(available)

    if total_available > 5:
        target_replica_count -= 1
    
    if total_available <= 0:
        target_replica_count += 1

    evaluation = {}
    evaluation["target_replicas"] = target_replica_count
    sys.stdout.write(json.dumps(evaluation))

if __name__ == "__main__":
    main()
