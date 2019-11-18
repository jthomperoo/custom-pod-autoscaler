# Scale on Tweet example
This is a novelty example, showing an autoscaler that queries the Twitter API searching for tweets with a specific hashtag. Counting the number of tweets that contain '👍' and the number of tweets that contain '👎' and sets the number of replicas to the difference between number of tweets containing 👍 and 👎.

## Overview
Trying out this example requires a kubernetes cluster to try it out on. This guide will assume you are using a Minikube cluster.

### Deploy an app for the Tweet Scaler to manage
You need to deploy an app for the CPA to manage, this example uses `hello-kubernetes`, which you can deploy with this command:  
`kubectl apply -f deployment.yaml`  

### Enable CPAs
Using this Tweet Scaler example requires Custom Pod Autoscalers to be enabled on your kubernetes cluster, [follow this guide to set up CPAs on your cluster](https://github.com/jthomperoo/custom-pod-autoscaler-operator#installation).  

### Set up a developer account for the Twitter API
1. Create a twitter account.
2. [Apply for Twitter developer account.](https://developer.twitter.com/en/apply-for-access.html)
3. Go to [the developer portal, then apps.](https://developer.twitter.com/en/apps)
4. Create a new app.
5. Open your new app details, then go to Keys and tokens.
6. Generate a new access token, take note of all of the keys (consumer API keys and access tokens).

### Build the Tweet scaler.
If you are using Minikube, use this command to set up Docker to point to the Minikube registry:  
`eval $(minikube docker-env)`  
Use this docker command to build the tweet scaler:  
`docker build -t simple-pod-metrics-python .`  

### Deploy the Tweet scaler
There is a Custom Pod Autoscaler YAML definition in this example, `cpa.yaml`, you just need to update some of the placholder values in it to the Twitter auth keys and whatever hashtag you want to watch.
```yaml
    - name: consumerKey
      value: <PUT YOUR CONSUMER KEY HERE>
    - name: consumerSecret
      value: <PUT YOUR CONSUMER SECRET HERE>
    - name: accessToken
      value: <PUT YOUR ACCESS TOKEN HERE>
    - name: accessTokenSecret
      value: <PUT YOUR ACCESS TOKEN SECRET HERE>
    - name: hashtag
      value: <HASHTAG TO WATCH>
```
Once you have updated these values, deploy the Tweet scaler with this command:  
`kubectl apply -f cpa.yaml`  
The tweet scaler should now be running on the cluster, managing the `hello-kubernetes` deployment and watching the hashtag you have specified.