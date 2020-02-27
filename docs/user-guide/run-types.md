# Run Types

The `runType` field is provided to metric gathering, evaluation and hook logic. This field allows user logic to determine the circumstances under which they are being called.  

The possible values for `runType` are:  

- `scaler` - Called as part of the scheduled autoscaler.
- `api` - Called as part of an API request.
- `api_dry_run` - Called as part of an API request, but marked as `dry_run`; this means that it should not affect state/will not result in an actual scale. 