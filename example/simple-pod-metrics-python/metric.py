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

# JSON piped into this script example:
# {
#     "apiVersion": "v1",
#     "kind": "Pod",
#     "metadata": {
#         "creationTimestamp": "2019-11-14T01:52:21Z",
#         "generateName": "flask-metric-869879868f-",
#         "labels": {
#             "app": "flask-metric",
#             "pod-template-hash": "869879868f"
#         },
#         "name": "flask-metric-869879868f-2cslm",
#         "namespace": "default",
#         "ownerReferences": [
#             {
#                 "apiVersion": "apps/v1",
#                 "blockOwnerDeletion": true,
#                 "controller": true,
#                 "kind": "ReplicaSet",
#                 "name": "flask-metric-869879868f",
#                 "uid": "2b028109-4793-4409-bd9c-a44d74da2fbc"
#             }
#         ],
#         "resourceVersion": "208999",
#         "selfLink": "/api/v1/namespaces/default/pods/flask-metric-869879868f-2cslm",
#         "uid": "5a0ab9a6-dccc-497d-8d41-11f0408740b3"
#     },
#     "spec": {
#         "containers": [
#             {
#                 "image": "flask-metric:latest",
#                 "imagePullPolicy": "Never",
#                 "name": "flask-metric",
#                 "ports": [
#                     {
#                         "containerPort": 5000,
#                         "protocol": "TCP"
#                     }
#                 ],
#                 "resources": {},
#                 "terminationMessagePath": "/dev/termination-log",
#                 "terminationMessagePolicy": "File",
#                 "volumeMounts": [
#                     {
#                         "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
#                         "name": "default-token-48k2f",
#                         "readOnly": true
#                     }
#                 ]
#             }
#         ],
#         "dnsPolicy": "ClusterFirst",
#         "enableServiceLinks": true,
#         "nodeName": "minikube",
#         "priority": 0,
#         "restartPolicy": "Always",
#         "schedulerName": "default-scheduler",
#         "securityContext": {},
#         "serviceAccount": "default",
#         "serviceAccountName": "default",
#         "terminationGracePeriodSeconds": 30,
#         "tolerations": [
#             {
#                 "effect": "NoExecute",
#                 "key": "node.kubernetes.io/not-ready",
#                 "operator": "Exists",
#                 "tolerationSeconds": 300
#             },
#             {
#                 "effect": "NoExecute",
#                 "key": "node.kubernetes.io/unreachable",
#                 "operator": "Exists",
#                 "tolerationSeconds": 300
#             }
#         ],
#         "volumes": [
#             {
#                 "name": "default-token-48k2f",
#                 "secret": {
#                     "defaultMode": 420,
#                     "secretName": "default-token-48k2f"
#                 }
#             }
#         ]
#     },
#     "status": {
#         "conditions": [
#             {
#                 "lastProbeTime": null,
#                 "lastTransitionTime": "2019-11-14T01:52:21Z",
#                 "status": "True",
#                 "type": "Initialized"
#             },
#             {
#                 "lastProbeTime": null,
#                 "lastTransitionTime": "2019-11-17T22:09:33Z",
#                 "status": "True",
#                 "type": "Ready"
#             },
#             {
#                 "lastProbeTime": null,
#                 "lastTransitionTime": "2019-11-17T22:09:33Z",
#                 "status": "True",
#                 "type": "ContainersReady"
#             },
#             {
#                 "lastProbeTime": null,
#                 "lastTransitionTime": "2019-11-14T01:52:21Z",
#                 "status": "True",
#                 "type": "PodScheduled"
#             }
#         ],
#         "containerStatuses": [
#             {
#                 "containerID": "docker://ebc2bf777301a37f6f28c253e3125c1522ca891d57de8810567b9768f66c8abb",
#                 "image": "flask-metric:latest",
#                 "imageID": "docker://sha256:b085c0b2703e02117834ae1fc6d640c54b33ebb2b198022a126cb5c0e4d71917",
#                 "lastState": {
#                     "terminated": {
#                         "containerID": "docker://fa7487b4f720e5fba02201f867afd05538f5fd21f2a25c93e235b83484cef21c",
#                         "exitCode": 255,
#                         "finishedAt": "2019-11-17T22:08:29Z",
#                         "reason": "Error",
#                         "startedAt": "2019-11-17T10:58:25Z"
#                     }
#                 },
#                 "name": "flask-metric",
#                 "ready": true,
#                 "restartCount": 7,
#                 "started": true,
#                 "state": {
#                     "running": {
#                         "startedAt": "2019-11-17T22:09:32Z"
#                     }
#                 }
#             }
#         ],
#         "hostIP": "10.0.2.15",
#         "phase": "Running",
#         "podIP": "172.17.0.2",
#         "podIPs": [
#             {
#                 "ip": "172.17.0.2"
#             }
#         ],
#         "qosClass": "BestEffort",
#         "startTime": "2019-11-14T01:52:21Z"
#     }
# }


def main():
    # Parse JSON into a dict
    pod = json.loads(sys.stdin.read())
    metric(pod)

def metric(pod):
    # Get Pod IP
    status = pod["status"]
    ip = status["podIP"]
    try:
        # Make request to Pod metric endpoint
        # (see ./app folder for simple flask app exposing this endpoint)
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
