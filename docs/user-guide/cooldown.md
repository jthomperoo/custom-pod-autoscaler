# Cooldown

A cooldown can be implemented by using the `downscaleStabilization` configuration option. This works in the same way that [the Horizontal Pod Autoscaler downscale stabilization](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-cooldown-delay) works.  

A stabilization window is simply defined by providing the window period in seconds; the autoscaler will keep track of any evaluations that result in a scale. When a scale is triggered, the autoscaler will look back at all evaluations that are within the stabilization period specified and pick out the one with the highest amount of replicas. The result of this is a more smoothed downscaling effect that does not limit upscaling, which can reduce thrashing.  

See the [configuration reference for more details](../../reference/configuration#downscalestabilization).