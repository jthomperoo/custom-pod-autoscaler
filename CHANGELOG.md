# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Changed
- The `command` for `shell` methods is now an array of arguments, rather than a string.

## [v0.9.0] - 2020-01-19
### Added
- Support for other entrypoints other than `/bin/sh`, can specify an entrypoint for the shell command method.
- Add logging library `glog` to allow logging at levels of severity and verbosity.
- Can specify verbosity level of logs via the `logVerbosity` configuration option.
### Changed
- Can scale ReplicaSets, ReplicationControllers and StatefulSets alongside Deployments.
- ResourceMetrics fields have `resourceName` and `resource` rather than `deploymentName` and `deployment`. In JSON this means that only the resource name will be exposed via field `resource`.
- Uses scaling API rather than manually adjusting replica count on resource.
- Matches using match selector rather than incorrectly using resource labels and building a different selector.

## [v0.8.0] - 2019-12-17
### Added
- New `startTime` configuration option in milliseconds; allows specifying a time that the interval should count up from when starting. This allows specifying a nearest time to start at, for example setting it to `60000` would start running at the closest minute, setting it to `15000` would start running at the closest 15 seconds e.g. :15 :30 :45.
- Support for JSON configuration, configuration file can now be in either YAML or JSON.
### Changed
- Replaced shell command with a generic method, allowing different methods to be supported. For example, instead of:
```yaml
evaluate: "python /evaluate.py"
evaluateTimeout: 2500
```
It is now:
```yaml
evaluate: 
  type: "shell"
  timeout: 2500
  shell: "python /evaluate.py"
```

## [v0.7.0] - 2019-12-08
### Added
- New `run_type` flag to the ResourceMetrics; allows scripts to understand the context of how it is being called.
    * Two values, either `api` triggered by an API call, or `scaler` which means it was triggered by the autoscaling logic.
### Changed
- Provide full metric information to be piped into the evaluation command, including the resource being managed.

## [0.6.0] - 2019-11-20
### Added
- Allow setting minimum and maximum replicas, with `minReplicas` and `maxReplicas` options - if the evaluation is above maxReplicas the resource is only scaled up to `maxReplicas` value, if the evaluation is below `minReplicas` the resource is only scaled down to `minReplicas`.
- Can disable autoscaling for a resource by setting its `replicas` to `0`.
### Changed
- The `target_replicas` field in an evaluation is no longer optional.

## [0.5.0] - 2019-11-18
### Added
- Allow specification of how metrics/evaluations should be run with `runMode`, either `per-pod` or `per-resource`. Mode `per-pod` means run the metric gathering command individually per pod, with pod info piped in. `per-resource` means run the metric gathering command per resource, with the resource info piped in. 
### Changed
- Metrics are now tied to a resource name, rather than a pod name - with `resource` rather than `pod` as the field in metrics, e.g.
```json
{
    "resource": "resource-name",
    "value:" "value"
}
```

## [0.4.0] - 2019-11-16
### Changed
- Path to config file specified using `configPath` rather than `config_path` - consistency with other config options.
- Not Found (404) and Method Not Allowed (405) now return valid JSON alongside the error code with a message explaining the error.
- Only allow management of a single deployment.
- Use `ScaleTargetRef` rather than `selector` for deciding which resources to manage, more consistent with Horizontal Pod Autoscaler.
- Simplified evaluation, now when hitting CPA API will just return the `target_replicas` rather than additional info and complicated JSON.

## [0.3.0] - 2019-11-03
### Added
- Two new configuration options, `metric_timeout` and `evaluate_timeout`. Allows timeouts to be set for `metric` command and `evaluate` commands; default `3000` milliseconds.
### Changed
- Read configuration in environment variables in consistent way with YAML, env vars are all lowercase.

## [0.2.0] - 2019-10-28
### Added
- Simple API for querying metrics/evaluations.
- Graceful shutdown of API and scaler.
- Namespace support for managing pods.
- Configuration of interval via environment variables/custom resource definition.
- Binary deployed with a release onto GitHub releases.

## [0.1.0] - 2019-09-30
### Added
- Allow specification of deployment to manage with a selector.
- Gather pods for managed deployment.
- Run user defined metric every set interval for each pod.
- Run user defined evaluation based on metric results.
- Updates target number of replicas for a deployment based on evaluation.
- Deploy image to Docker Hub.

[Unreleased]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/v0.9.0...HEAD
[v0.9.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.6.0...v0.7.0
[0.6.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.5.0...0.6.0
[0.5.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.4.0...0.5.0
[0.4.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.3.0...0.4.0
[0.3.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.2.0...0.3.0
[0.2.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.1.0...0.2.0
[0.1.0]: https://github.com/jthomperoo/custom-pod-autoscaler/releases/tag/0.1.0