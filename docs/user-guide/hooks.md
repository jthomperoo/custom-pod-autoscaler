# Hooks

Hooks provide a way to inject user logic to different points in the autoscaler execution process through the [use of
`methods`](../methods). Each  hook is provided with different data.

* `preMetric` - Runs before metric gathering, given metric gathering input.
* `postMetric` - Runs after metric gathering, given metric gathering input and result.
* `preEvaluate` - Runs before evaluation, given evaluation input.
* `postEvaluate` - Runs after evaluation, given evaluation input and result.
* `preScale` - Runs before scaling decision, given min and max replicas, current replicas, target replicas, and
resource being scaled.
* `postScale` - Runs after scaling decision, given min and max replicas, current replicas, target replicas, and
resource being scaled.

For more detailed information on values provided into these hooks and when they are called, [check out the
configuration reference](../../reference/configuration).
