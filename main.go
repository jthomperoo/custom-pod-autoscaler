/*
Copyright 2019 The Custom Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/models"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
)

func main() {
	// creates the in-cluster config
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		panic(err.Error())
	}

	dynamicKubeConfig, err := config.LoadConfig()
	if err != nil {
		panic(err.Error())
	}

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	ticker := time.NewTicker(time.Duration(dynamicKubeConfig.Interval) * time.Millisecond)
	quit := make(chan struct{})
	go evaluate(clientset, deploymentsClient, dynamicKubeConfig, ticker, quit)

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Placeholder")
}

func executeShellWithValuePipe(command string, value string) (*bytes.Buffer, error) {
	// Build command string with value piped into it
	commandString := fmt.Sprintf("echo '%s' | %s", value, command)
	cmd := exec.Command("/bin/sh", "-c", commandString)
	// Set up byte buffers to read stdout and stderr
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		// Output stderr
		println(string(errb.String()))
		return nil, err
	}
	return &outb, nil
}

func evaluate(clientset *kubernetes.Clientset, deploymentsClient v1.DeploymentInterface, dynamicKubeConfig *config.Config, ticker *time.Ticker, quit chan struct{}) {
	for {
		select {
		case <-ticker.C:
			deployments, err := deploymentsClient.List(metav1.ListOptions{LabelSelector: dynamicKubeConfig.Selector})
			if err != nil {
				panic(err.Error())
			}
			for _, deployment := range deployments.Items {
				labels := deployment.GetLabels()
				pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", labels["app"])})
				if err != nil {
					panic(err.Error())
				}
				var metrics []*models.Metric
				for _, pod := range pods.Items {
					podJSON, err := json.Marshal(pod)
					if err != nil {
						println(err.Error())
						continue
					}
					outb, err := executeShellWithValuePipe(dynamicKubeConfig.Metric, string(podJSON))
					if err != nil {
						println(err.Error())
						continue
					}
					metrics = append(metrics, &models.Metric{
						Pod:   pod.GetName(),
						Value: string(outb.String()),
					})
				}
				metricJSON, err := json.Marshal(metrics)
				if err != nil {
					println(err.Error())
				}
				outb, err := executeShellWithValuePipe(dynamicKubeConfig.Evaluate, string(metricJSON))
				if err != nil {
					println(err.Error())
					continue
				}
				evaluation := &models.Evaluation{}
				json.Unmarshal(outb.Bytes(), evaluation)
				if evaluation.TargetReplicas != *deployment.Spec.Replicas {
					deployment.Spec.Replicas = &evaluation.TargetReplicas
					_, err = deploymentsClient.Update(&deployment)
					if err != nil {
						println(err.Error())
					}
				}
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
