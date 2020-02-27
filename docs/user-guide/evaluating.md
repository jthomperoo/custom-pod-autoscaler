# Evaluating

The evaluation stage is the second and final stage of the autoscaler. This stage is responsible for 
taking all the metrics gathered by the [metric gathering stage](../metric-gathering) and using them 
to make a decision; how many replicas should the managed resource have. The autoscaler base program 
provides the metrics gathered by the previous stage, wrapped in JSON with some additional 
information, the evaluator is expected to take this and return a JSON response describing how 
many replicas the resource should have, or if errors occur these errors are to be returned and 
error information surfaced.

# Information in

The evaluation stage recieves the output of the [metric gathering stage](../metric-gathering) stage 
passed into it wrapped in JSON, with additional information such as the run type and the resource the metric was gathered from.  
An example of the JSON passed into the evaluation stage:
```json
{
  "resourceMetrics": {
    "metrics": [
      {
        "resource": "hello-kubernetes",
        "value": "3"
      }
    ],
    "resource": {
      "kind": "Deployment",
      "apiVersion": "apps/v1",
      "metadata": {
        "name": "hello-kubernetes",
        "namespace": "default",
        "labels": {
          "numPods": "3"
        },
      },
  },
  "runType": "scaler"
}
```
How the information is provided depends on the method used, e.g. if it is a shell
command method the information will be provided through standard in and piped to the script, see
the [methods section for more information](../methods).

# Information out

The evaluation stage, if successful, is expected to calculate the replica count and return it 
in a wrapped JSON object, for example:
```json
{
  "targetReplicas": 5
}
```

How the evaluator returns information/errors is dependent on the method used for the evaluator, 
for example if it is a shell command method, information is returned by writing to 
standard out, and an error is signified by a non-zero exit code - with further error information 
stored to standard error and standard out, see the [methods section for more information](../methods).
