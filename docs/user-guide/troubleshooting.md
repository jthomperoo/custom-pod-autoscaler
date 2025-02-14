# Troubleshooting

This page explains some of the common issues people may experience.

## Autoscaler is Forbidden From Managing Custom Resources

You may encounter an error like this:

`
E0214 11:30:09.556332       1 main.go:266] Error while autoscaling: failed to get managed resource: logstashes.logstash.k8s.elastic.co "quickstart" is forbidden: User "system:serviceaccount:default:python-custom-autoscaler" cannot get resource "logstashes" in API group "logstash.k8s.elastic.co" in the namespace "default"
`

And see that your autoscaler is not scaling the target resource if you are targeting a custom resource.

This is because your autoscaler does not have the correct permissions to manage the resource you are
targeting.

See [the Custom Resources page for details on how to resolve this](./custom-resources.md).
