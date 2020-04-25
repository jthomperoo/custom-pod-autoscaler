# Scaling To and From Zero

Custom Pod Autoscalers can scale to and from zero; it is dependent on the
`minReplicas` configuration option.

## Disable Scale to Zero

If the `minReplicas` configuration option is set to anything but `0` scaling to
zero is disabled; if the replica count of a resource being managed is set to `0`
it will be treated as autoscaling disabled and the CPA will not scale it up or
down.

## Enable Scale to Zero

If the `minReplicas` configuration option is set to `0` scaling to and from `0`
is enabled, allowing the autoscaler to set the replica count to `0` and then
back up again.