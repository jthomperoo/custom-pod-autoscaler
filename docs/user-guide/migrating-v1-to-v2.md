# Migrating v1 to v2

The breaking changes when moving from Custom Pod Autoscaler `v1` to `v2` are:

- The Go package structure has been adjusted.
- The Go data structs have been renamed.

Therefore if your autoscaler does not rely directly on the Go code in this project then you can safely upgrade without
any issues.

If your Go code does rely on this project, the breaking changes to the Go codebase are:

The Go package that should be imported is `github.com/jthomperoo/custom-pod-autoscaler/v2` instead of
`github.com/jthomperoo/custom-pod-autoscaler`.

- Packages no longer exported:
  - `autoscaler`
  - `execute`
  - `execute/shell`
  - `execute/http`
  - `fake`
  - `resourceclient`
- Any functions or interfaces in the following packages are no longer exported:
  - `api/v1`
  - `config`
  - `evaluate`
  - `metric`
  - `scale`
- The following structs are no longer exported:
  - `api/v1`
    - `API`
  - `evaluate`
    - `Evaluator`
  - `metric`
    - `Gatherer`
  - `scale`
    - `Scale`
    - `TimestampedEvaluation`
- The following structs have been renamed:
  - `evaluate`
    - `Spec` -> `Info`
  - `metric`
    - `Metric` -> `ResourceMetric`
    - `Spec` -> `Info`
  - `scale`
    - `Spec` -> `Info`
- The following constants have been moved and renamed:
  - `api`
    - `RunType` -> `config.APIRunType`
    - `RunTypeDryRun` -> `config.APIDryRunRunType`
  - `autoscaler`
     - `RunType` -> `config.ScalerRunType`
- The `config` package now has a new function `NewConfig` which returns an instance of the config with the default
values set.
