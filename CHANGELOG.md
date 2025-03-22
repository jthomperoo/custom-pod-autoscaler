# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic
Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v2.12.2] - 2025-03-22
### Changed
- Upgraded package dependencies.
### Added
- New binary build targets:
  - Linux i386 (x86)
  - Linux ARM
  - Linux ARM64
  - Darwin AMD64
  - Darwin ARM64
  - Windows i386 (x86)
  - Windows AMD64
- New Docker platform:
  - Linux ARM64

## [v2.12.1] - 2025-02-14
### Fixed
- Handling of plurals is more robust, it no longer simply adds a 's' to the end which failed for plurals such
as 'logstash' -> 'logstashes', now queries using the discovery client and REST mapper.

## [v2.12.0] - 2025-02-12
### Added
- Added support for Python 3.13 Docker image (`custompodautoscaler/python-3-13`).
### Changed
- Upgraded to Go `v1.24`.
- Upgraded package dependencies.
- Updated `custompodautoscaler/python` to track Python 3.13.
- Updated `custompodautoscaler/python-3-12` to track the latest Debian stable version.
### Removed
- Dropped support for Python 3.8 Docker image (`custompodautoscaler/python-3-8`).

## [v2.11.0] - 2024-03-22
### Changed
- Updated to `k8shorizmetrics/v3`.
- Partial K8s metrics server errors are logged to help with debugging.

## [v2.10.0] - 2023-12-24
### Added
- Added Python 3.12 Docker image (`custompodautoscaler/python-3-12`).
### Removed
- Dropped support for Python 3.6 and 3.7 Docker images (`custompodautoscaler/python-3-6` and
`custompodautoscaler/python-3-7`) which are EOL ([see Python version lifecycles
here](https://devguide.python.org/versions/)).
- Dropped support for OpenJDK image (`custompodautoscaler/openjdk-11`) since the base images are EOL ([see deprecation
notice here](https://hub.docker.com/_/openjdk)). Instead use a Custom Docker Image and install a JRE into it, [see
the updated simple-pod-metrics-java example for
details](https://github.com/jthomperoo/custom-pod-autoscaler/tree/master/example/simple-pod-metrics-java).
### Changed
- Updated `custompodautoscaler/python` to track Python 3.12.
- Updated package dependencies.

## [v2.9.0] - 2023-12-23
### Changed
- Switched from performing UPDATE using the K8s API to using PATCH, avoids race condition if resource is modified
while calculating the target number of replicas. ([#187](https://github.com/jthomperoo/custom-pod-autoscaler/pull/187))

## [v2.8.0] - 2023-08-15
### Changed
- Upgraded to Go `v1.21`.
- Upgraded package dependencies.
- Dropped support for Kubernetes `v1.23`, upgraded to use the `autoscaling/v2` API.

## [v2.7.0] - 2022-05-14
### Changed
- Upgraded K8s client libraries to `v0.24.0`.
- Upgraded to Go `v1.17`.
- Extracted out internal K8s metrics calculations to use
[jthomperoo/k8shorizmetrics](https://github.com/jthomperoo/k8shorizmetrics) instead.

## [v2.6.0] - 2022-04-17
### Added
- Mutual TLS support for `http` methods.
  - Three new fields to the `http` method, `caCert`, `clientCert` and `clientKey` which allow paths to be specified to
  certs and keys to support mutual TLS and client certificates.

## [v2.5.0] - 2022-03-26
### Changed
- Now internally using the Kubernetes Scale API for determining replica count and selectors, removing any direct
dependencies on the Kubernetes core types when used in conjunction with the Kubernetes unstructured API.
- Can now install the project by using the `go install` command since replace directives have been removed.

## [v2.4.1] - 2022-02-22
### Fixed
- K8s metrics now use the Spec.Selector to select resources from the metrics server, rather than incorrectly using
the labels of the resource being scaled.
- K8s metrics now support [match
expressions](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#resources-that-support-set-based-requirements)
where available (`Deployment`, `ReplicaSet`, `StatefulSet`, `Rollout`).

## [v2.4.0] - 2021-12-28
### Changed
- Error messages now include more context about when they occured, rather than simply outputting the minimum error
information.

## [v2.3.0] - 2021-12-27
### Added
- Support for `argoproj.io/v1alpha1` `Rollout` resource.
### Changed
- Add clarity between `autoscaling/v1` and `autoscaling/v2beta2` APIs, no breaking changes.
### Security
- Updated k8s client libraries to `v0.21.8` to resolve
[GHSA-f5f7-6478-qm6p](https://github.com/advisories/GHSA-f5f7-6478-qm6p).

## [v2.2.0] - 2021-08-14
### Added
- New `openjdk-11` image which includes Java 11 runtime.

## [v2.1.0] - 2021-06-16
### Added
- New `logStderr` option for `shell` methods, intended to help debugging by printing out the stderr of a successful
`shell` method execution - default value `false`.
### Changed
- Error returned when parsing invalid JSON for an evaluation improved to help debugging.

## [v2.0.0] - 2021-04-15
### Changed
- **BREAKING CHANGE** Project's Go code restructured, limited exposed packages. See [the migration
guide](./docs/user-guide/migrating-v1-to-v2.md) for full details.
- **BREAKING CHANGE** Pre-scaling hook moved to after the downscale stabilization value has been calculated.
- `k8smetric` package now exposed to allow easy Go dependency marshal/unmarshal of K8s metrics.
### Fixed
- `targetReplicas` value now set properly, set to the pre-stabilized target replica value.
### Added
- The Custom Pod Autoscaler version is now printed to the log on autoscaler startup.

## [v1.1.0] - 2021-04-08
### Added
- Feature to provide standard K8s metrics to the metric gathering stage of the autoscaler
  - Can now provide a list of Metric Specs (similar to the Horizontal Pod Autoscaler) to choose which metrics to
  include in the data sent to the metric gathering stage with the `kubernetesMetricSpecs` configuration option.
  - Can provide the `requireKubernetesMetrics` option to fail if the metrics server query fails.
  - Can provide `initialReadinessDelay` and `cpuInitializationPeriod` values for use when querying the metrics server.
### Changed
- Switched docs theme to material theme.

## [v1.0.1] - 2020-09-12
### Added
- Three new Python images:
  * `custompodautoscaler/python-3-6` tracks latest stable Python 3.6.x release.
  * `custompodautoscaler/python-3-7` tracks latest stable Python 3.7.x release.
  * `custompodautoscaler/python-3-8` tracks latest stable Python 3.8.x release.
### Changed
- The `custompodautoscaler/python` image now tracks the latest stable Python 3.x release.

## [v1.0.0] - 2020-07-19

## [v0.13.0] - 2020-07-18
### Added
- HTTP method, allows specifying an HTTP request to make as a method - for example querying an external API as part of
the metric gathering stage.
- Extra error checking for `shell` method, will no longer throw nil pointer if no shell configuration is provided, more
useful error is raised instead.

## [v0.12.0] - 2020-04-25
### Changed
- Support scaling to and from zero, matching misimplemented functionality from Horizontal Pod Autoscaler.

## [v0.11.0] - 2020-02-28
### Added
- Series of hooks for injecting user logic throughout the execution process.
  * `preMetric` - Runs before metric gathering, given metric gathering input.
  * `postMetric` - Runs after metric gathering, given metric gathering input and result.
  * `preEvaluate` - Runs before evaluation, given evaluation input.
  * `postEvaluate` - Runs after evaluation, given evaluation input and result.
  * `preScale` - Runs before scaling decision, given min and max replicas, current replicas, target replicas, and
  resource being scaled.
  * `postScale` - Runs before scaling decision, given min and max replicas, current replicas, target replicas, and
  resource being scaled.
- New `downscaleStabilization` option, based on [the Horizontal Pod Autoscaler downscale
  stabilization](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-cooldown-delay),
  operates by taking the maximum target replica count over the stabilization window.
### Changed
- Metrics from API now returns the entire resource definition as JSON rather than just the resource name.
- Changed JSON generated to be in `camelCase` rather than `snake_case` for consistency with the Kubernetes API.
  * Evaluation now uses `targetReplicas` over `target_replicas`.
  * ResourceMetric now uses `runType` over `run_type`.
  * Scale hook now provided with `minReplicas`, `maxReplicas`, `currentReplicas` and `targetReplicas` rather than their
  snakecase equivalents.
- Metric gathering and hooks have access to `dryRun` field, allowing them to determine if they are called as part of a
  dry run.
- Standardised input to metric gatherer, evaluator and scaler to take specs rather than lists of parameters, allowing
  easier serialisation for hooks.
- Endpoint `/api/v1/metrics` now accepts the optional `dry_run` parameter for marking metric gathering as in dry run
  mode.
- `ResourceMetrics` replaced with a list of `Metric` and a `Resource`.
- `/api/v1/metrics` now simply returns a list of `Metrics` rather than a `ResourceMetrics`.

### Removed
- `ResourceMetrics` struct removed as it was redundant.

## [v0.10.0] - 2020-01-22
### Added
- Set up API to be versioned, starting with `v1`.
- Can now manually trigger scaling through the API.
- Added extra `run_type` flag, `api_dry_run`, for evaluations through the API in `dry_run` mode.
- Added `apiConfig` to hold configuration for the REST API.
- Added extra configuration options within `apiConfig`.
  * `enabled` - allows enabling or disabling the API, default enabled (`true`).
  * `useHTTPS` - allows enabling or disabling HTTPS for the API, default off (`false`).
  * `certFile` - cert file to be used if HTTPS is enabled.
  * `keyFile` - key file to be used if HTTPS is enabled.

### Changed
- The `command` for `shell` methods is now an array of arguments, rather than a string.
- The `/api/v1/evaluation` endpoint now requires `POST` rather than `GET`.
- The `/api/v1/evaluation` endpoint now accepts an optional parameter, `dry_run`. If `dry_run` is true the evaluation
will be retrieved in a read-only manner, the scaling will not occur. If it is false, or not provided, the evaluation
will be retrieved and then used to apply scaling to the target.
- Moved `port` and `host` configuration options into the `apiConfig` settings.

## [v0.9.0] - 2020-01-19
### Added
- Support for other entrypoints other than `/bin/sh`, can specify an entrypoint for the shell command method.
- Add logging library `glog` to allow logging at levels of severity and verbosity.
- Can specify verbosity level of logs via the `logVerbosity` configuration option.
### Changed
- Can scale ReplicaSets, ReplicationControllers and StatefulSets alongside Deployments.
- ResourceMetrics fields have `resourceName` and `resource` rather than `deploymentName` and `deployment`. In JSON this
  means that only the resource name will be exposed via field `resource`.
- Uses scaling API rather than manually adjusting replica count on resource.
- Matches using match selector rather than incorrectly using resource labels and building a different selector.

## [v0.8.0] - 2019-12-17
### Added
- New `startTime` configuration option in milliseconds; allows specifying a time that the interval should count up from
when starting. This allows specifying a nearest time to start at, for example setting it to `60000` would start running
at the closest minute, setting it to `15000` would start running at the closest 15 seconds e.g. :15 :30 :45.
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
    * Two values, either `api` triggered by an API call, or `scaler` which means it was triggered by the autoscaling
    logic.
### Changed
- Provide full metric information to be piped into the evaluation command, including the resource being managed.

## [0.6.0] - 2019-11-20
### Added
- Allow setting minimum and maximum replicas, with `minReplicas` and `maxReplicas` options - if the evaluation is above
maxReplicas the resource is only scaled up to `maxReplicas` value, if the evaluation is below `minReplicas` the
resource is only scaled down to `minReplicas`.
- Can disable autoscaling for a resource by setting its `replicas` to `0`.
### Changed
- The `target_replicas` field in an evaluation is no longer optional.

## [0.5.0] - 2019-11-18
### Added
- Allow specification of how metrics/evaluations should be run with `runMode`, either `per-pod` or `per-resource`. Mode
`per-pod` means run the metric gathering command individually per pod, with pod info piped in. `per-resource` means run
the metric gathering command per resource, with the resource info piped in.
### Changed
- Metrics are now tied to a resource name, rather than a pod name - with `resource` rather than `pod` as the field in
metrics, e.g.
```json
{
    "resource": "resource-name",
    "value:" "value"
}
```

## [0.4.0] - 2019-11-16
### Changed
- Path to config file specified using `configPath` rather than `config_path` - consistency with other config options.
- Not Found (404) and Method Not Allowed (405) now return valid JSON alongside the error code with a message explaining
the error.
- Only allow management of a single deployment.
- Use `ScaleTargetRef` rather than `selector` for deciding which resources to manage, more consistent with Horizontal
Pod Autoscaler.
- Simplified evaluation, now when hitting CPA API will just return the `target_replicas` rather than additional info
and complicated JSON.

## [0.3.0] - 2019-11-03
### Added
- Two new configuration options, `metric_timeout` and `evaluate_timeout`. Allows timeouts to be set for `metric`
command and `evaluate` commands; default `3000` milliseconds.
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

[Unreleased]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.12.2...HEAD
[v2.12.2]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.12.1...v2.12.2
[v2.12.1]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.12.0...v2.12.1
[v2.12.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.11.0...v2.12.0
[v2.11.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.10.0...v2.11.0
[v2.10.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.9.0...v2.10.0
[v2.9.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.8.0...v2.9.0
[v2.8.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.7.0...v2.8.0
[v2.7.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.6.0...v2.7.0
[v2.6.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.5.0...v2.6.0
[v2.5.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.4.1...v2.5.0
[v2.4.1]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.4.0...v2.4.1
[v2.4.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.3.0...v2.4.0
[v2.3.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.2.0...v2.3.0
[v2.2.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.1.0...v2.2.0
[v2.1.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v2.0.0...v2.1.0
[v2.0.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v1.1.0...v2.0.0
[v1.1.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v1.0.1...v1.1.0
[v1.0.1]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v1.0.0...v1.0.1
[v1.0.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v0.13.0...v1.0.0
[v0.13.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v0.12.0...v0.13.0
[v0.12.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v0.11.0...v0.12.0
[v0.11.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v0.10.0...v0.11.0
[v0.10.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v0.9.0...v0.10.0
[v0.9.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v0.8.0...v0.9.0
[v0.8.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/v0.7.0...v0.8.0
[v0.7.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.6.0...v0.7.0
[0.6.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.5.0...0.6.0
[0.5.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.4.0...0.5.0
[0.4.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.3.0...0.4.0
[0.3.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.2.0...0.3.0
[0.2.0]:
https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.1.0...0.2.0
[0.1.0]: https://github.com/jthomperoo/custom-pod-autoscaler/releases/tag/0.1.0
