# Methods

Methods specify how user logic should be called by the Custom Pod Autoscaler base program. Methods 
can be specified for metric gathering and evaluating.

> Note: as of `v0.8.0` only the `shell` method is supported.

# shell

The shell method allows specifying a shell command, run through `/bin/sh`. Any relevant 
information will be provided to the command specified by piping the information in through standard 
in. Data is returned by writing to standard out. An error is signified by exiting with a non-zero 
exit code; if an error occurs the autoscaler will capture all standard error and out and log it.  

This is an example configuration of the shell method for the metric gatherer:
```yaml
metric: 
  type: "shell"
  timeout: 2500
  shell: "python /metric.py"
```
Breaking this example down:

- `type` = the type of the method, for this example it is a `shell` method.
- `timeout` = the maximum time the method can take in milliseconds, for this example it is `2500` (2.5 seconds), if it takes longer than this it will count the method as failing.
- `shell` = the shell command to execute for this method.

This is a metric configuration that will always fail:
```yaml
metric: 
  type: "shell"
  timeout: 2500
  shell: "exit 1"
```

This is a metric configuration that will return `5` as a metric.
```yaml
metric: 
  type: "shell"
  timeout: 2500
  shell: "echo '5'"
```