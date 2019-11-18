# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.5.0...HEAD
[0.5.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.4.0...0.5.0
[0.4.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.3.0...0.4.0
[0.3.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.2.0...0.3.0
[0.2.0]: https://github.com/jthomperoo/custom-pod-autoscaler/compare/0.1.0...0.2.0
[0.1.0]: https://github.com/jthomperoo/custom-pod-autoscaler/releases/tag/0.1.0