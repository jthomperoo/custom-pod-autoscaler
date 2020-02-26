# Post Scale Hook

This is an autoscaler that makes use of one of the available hooks, the `postScale` hook to run a script after a scale occurs. This is based on the [getting started tutorial example autoscaler](../python-custom-autoscaler), augmented with a new hook.

## Overview

This autoscaler is functionally the same as the [getting started tutorial example autoscaler](../python-custom-autoscaler), but with a new `postScale` hook.  
The `postScale` hook is defined in `config.yaml`:
```yaml
postScale:
  type: "shell"
  timeout: 2500
  shell: 
    entrypoint: "python"
    command: 
      - "/post_scale.py"
```
This will run the `post_scale.py` Python script after scaling.  

The `post_scale.py` is a simple script that dumps whatever is piped to it into a JSON file.
```python
import os
import json
import sys

def main():
    # Parse scale info JSON into a dict
    scale_info = json.loads(sys.stdin.read())

    with open("/post_scale_data.json", "w+") as file:
        file.write(json.dumps(scale_info))

if __name__ == "__main__":
    main()

```

When running the autoscaler, check the `post_scale_data.json` to see the JSON data that is passed to the hook and to check the hook is working.  
The autoscaler is running with `logVerbosity` set to `3`, meaning that if you look at the logs it should give you detailed information; including if the hook is executed.

## Usage

### Enable CPAs
Using this CPA requires CPAs to be enabled on your kubernetes cluster, [follow this guide to set up CPAs on your cluster](https://github.com/jthomperoo/custom-pod-autoscaler-operator#installation).  

### Switch to target the Minikube registry
Target the Minikube registry for building the image:  
`eval $(minikube docker-env)`

### Deploy an app for the CPA to manage
You need to deploy an app for the CPA to manage:  
* Deploy the app using a deployment.  
`kubectl apply -f deployment.yaml`  
Now you have an app running to manage scaling for.

### Build CPA image
Once CPAs have been enabled on your cluster, you need to build this example, run these commands to build the example:  
* Build the example image.  
`docker build -t post-scale-hook .`  
* Deploy the CPA using the image just built.  
`kubectl apply -f cpa.yaml`  
Now the CPA should be running on your cluster, managing the app we previously deployed.

## Conclusion

Have a look at [the wiki for more information, such as guides and references](https://custom-pod-autoscaler.readthedocs.io/en/latest/)