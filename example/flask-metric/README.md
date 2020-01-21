# Simple Flask app
This is an example application that is used in a number of the example custom pod autoscalers:

* [Simple Pod Metrics in Python](../simple-pod-metrics-python/README.md)
* [Simple Pod Metrics in Golang](../simple-pod-metrics-golang/README.md)


## Overview
This application will be scaled up and down, managed by a Custom Pod Autoscaler.  
This app exposes an endpoint, `GET /metric` - this returns a JSON metric that looks like this example:
```json
{
    "value": 3,
    "available": 2,
    "min": 0,
    "max": 5
}
```  
Value is an arbitrary value between the `min` and the `max`.  
Available is how far the value is from the max, i.e. `max - value`.  
Min and max are the minimum and maximum values for the `value`.  
This value can be incremented and decremented with `POST /increment` and `POST /decrement` endpoints respectively.  
There are three aliases set up to make this easy on the app Docker image:
* `metric` reports the current metric.
* `increment` increments the value.
* `decrement` decrements the value.
You can exec into the example app pod and increase/decrease the value and see how the CPA creates/deletes pods.  