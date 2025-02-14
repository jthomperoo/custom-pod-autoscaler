# Troubleshooting

This page explains some of the common issues people may experience.

## Autoscaler is Forbidden From Managing Custom Resources

You may encounter an error like this:

`
E0214 11:30:09.556332       1 main.go:266] Error while autoscaling: failed to get managed resource: logstashes.logstash.k8s.elastic.co "quickstart" is forbidden: User "system:serviceaccount:default:python-custom-autoscaler" cannot get resource "logstashes" in API group "logstash.k8s.elastic.co" in the namespace "default"
`

And see that your autoscaler is not scaling the target resource if you are targeting a custom resource.

This is because your autoscaler does not have the correct permissions to manage the resource you are
targeting. By default the [Custom Pod Autoscaler
Operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator) will provision a role for your autoscaler
which allows managing the built-in Kubernetes resources (deployments, replicasets, statefulsets,
replicationcontrollers), but if you are targeting a custom resource you need to provide your own role with the
correct permissions.

You need to tell the operator not to provision a role for you, and provide your own with the permissions needed,
for example to target Loadstash:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: python-custom-autoscaler
rules:
- apiGroups:
    - logstash.k8s.elastic.co
  resources:
    - logstashes
    - logstashes/scale
  verbs:
    - '*'
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: python-custom-autoscaler
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: python-custom-autoscaler
subjects:
- kind: ServiceAccount
  name: python-custom-autoscaler
roleRef:
  kind: Role
  name: python-custom-autoscaler
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: custompodautoscaler.com/v1
kind: CustomPodAutoscaler
metadata:
  name: python-custom-autoscaler
spec:
  template:
    spec:
      serviceAccountName: python-custom-autoscaler
      containers:
      - name: python-custom-autoscaler
        image: python-custom-autoscaler:latest
        imagePullPolicy: IfNotPresent
  scaleTargetRef:
    apiVersion: logstash.k8s.elastic.co/v1alpha1
    kind: Logstash
    name: quickstart
  provisionRole: false
  provisionRoleBinding: false
  provisionServiceAccount: false
  config:
    - name: interval
      value: "10000"
```

This takes over provisioning of the role, the role binding, and the service account from the operator.

For any custom resource the CPO can support scaling it if the resource implements the scale subresource, and if so
the permissions needed are generally:

```yaml
- apiGroups:
    - my.api.group
  resources:
    - mycustomresource
    - mycustomresource/scale
  verbs:
    - '*'
```

There is a special case for [Argo Rollouts](https://argoproj.github.io/rollouts/), which can simply provide the
`roleRequiresArgoRollouts` field. [See more information
here](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/v1.4.2/USAGE.md#automatically-provisioning-a-role-that-supports-argo-rollouts).

[See how to skip automatic role/any other resource provision
here](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/v1.4.2/USAGE.md#using-custom-resources).
