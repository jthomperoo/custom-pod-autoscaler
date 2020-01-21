# Configuration

Configuration is a key part of the Custom Pod Autoscaler, defining how to call user logic, alongside more fine tuned configuration options that the end user might adjust; such as polling interval.

## How to provide configuration

### Configuration file in image
Example:  
```yaml
evaluate: 
  type: "shell"
  timeout: 2500
  shell: 
    entrypoint: "python"
    command: 
      - "/evaluate.py"
metric: 
  type: "shell"
  timeout: 2500
  shell: 
    entrypoint: "python"
    command: 
      - "/metric.py"
runMode: "per-resource"
```

Configuration may be set by using a configuration file (default path `/config.yaml`), this can be baked into a Docker image to provide standard configuration for the autoscaler - for example metric gathering and evaluation methods.

### Configuration passed as environment variables (defined in deployment YAML)
Example:  
```yaml
  config: 
    - name: interval
      value: "10000"
    - name: startTime
      value: "60000"
```

Configuration may be passed as environment variables; these can be customised at deploy time rather than baked in at build time, so allow for more fine tuned customisation. The main way to define the environment variables should be using the `config` YAML description, which allows for key-value pairs to define each configuration option and the value it should have. All configuration options defined as key-value pairs in the `config` YAML are converted into environment variables by the operator; this allows autoscalers to extend the configuration options and use this `config` YAML to define configuration for the user logic. 

> Note: Configuration set as an environment variable takes precedence over configuration set in a configuration file, this allows the configuration file to act possibly as a set of defaults that can be overridden at deploy time.

## Methods
Example:  
```yaml
metric: 
  type: "shell"
  timeout: 2500
  shell: 
    entrypoint: "python"
    command: 
      - "/metric.py"
```
A method defines a hook for calling user logic, with a timeout to handle user logic that hangs.  

### type
- `shell` = call the user logic through a shell command, with the relevant information piped through to the command. The user logic communicates back with the autoscaler through exit codes and standard error and out. A non zero exit code tells the autoscaler that the user logic has failed; and the autoscaler will read in standard error and log it for debug purposes. If no error occurs, the autoscaler may read in the standard out and use it, e.g. for metric gathering.

Defines the type of the method.
### timeout
Defines how long the autoscaler should wait for the user logic to finish, if it exceeds this time it will assume the operation has failed and provide a timeout error.

### shell
Defines a shell method, which is a simple string with the shell command to execute. Shell commands executed through `/bin/sh`, with values piped in through standard in.


## configPath
```yaml
config: 
  - name: configPath
    value: "/config.yaml"
```
Default: `/config.yaml`  
This defines the path to the configuration file. Should only be defined as an environment variable (through deployment YAML), as defining the path to the configuration file inside the configuration file does not make sense.

## interval
Example:  
```yaml
interval: 15000
```
Default value: `15000`  
This defines in milliseconds how frequently the autoscaler should run. The autoscaler will run, then wait this many milliseconds before running again.
## evaluate
Example:  
```yaml
evaluate: 
  type: "shell"
  timeout: 2500
  shell: 
    entrypoint: "python"
    command: 
      - "/evaluate.py"
```
No default value, required to be set.  
This defines the evaluation logic that should be run, and how it should be triggered.  
[This is a `method`, see methods section for full configuration options of a method](#methods). 
## metric
Example:  
```yaml
metric: 
  type: "shell"
  timeout: 2500
  shell: 
    entrypoint: "python"
    command: 
      - "/metric.py"
```
No default value, required to be set.  
This defines the metric logic that should be run, and how it should be triggered.  
[This is a `method`, see methods section for full configuration options of a method](#methods).
## host
Example:  
```yaml
host: "0.0.0.0"
```
Default value: `0.0.0.0`  
Defines the host to use for hosting the Custom Pod Autoscaler API.
## port
Example:  
```yaml
port: "5000"
```
Default value: `5000`  
Defines the port to use for hosting the Custom Pod Autoscaler API.
## namespace
Example:  
```yaml
namespace: "default"
```
Default value: `default`  
Defines the namespace to look in for the resource being managed.
## minReplicas
Example:  
```yaml
minReplicas: 1
```
Default value: `1`  
Defines the minimum number of replicas allowed, resource won't be scaled below this value.
## maxReplicas
Example:  
```yaml
maxReplicas: 10
```
Default value: `10`  
Defines the maximum number of replicas allowed, resource won't be scaled above this value.
## runMode
Example:  
```yaml
runMode: "per-pod"
```
Default value: `per-pod`  

- `per-pod` = runs metric gathering per pod, individually running the user logic for each pod in the resource being managed, with the pod information provided to the user logic.
- `per-resource` = runs metric gathering per resource, running the user logic for only the resource being managed, with the resource information provided to the user logic.

Defines how the autoscaler runs the metric gathering user logic, changing the values that are provided to the metric gathering user logic and changing how frequently the metric gathering user logic is called.
## startTime
Example:  
```yaml
startTime: 1
```
Default value: `1`  
This defines in milliseconds a starting point for the scaler, with the scaler running as if it started at the time provided. Allows specifying that the autoscaler must start on a multiple of the interval from the start time. For example, a startTime of `60000` would result in the autoscaler starting at the next full minute. The default value will start the autoscaler after a single millisecond, close to instantly.

> Note: The scaling will not actually start until after one interval has passed.
## logVerbosity
Example:
```yaml
logVerbosity: 0
```
Default value: `0`  
This defines the verbosity of the logging, allowing for debugging errors/issues. Logging will occur for all values ABOVE and including the verbosity level set.  
Log levels:

* `0` - normal.
* `1` - verbose.
* `2` - more verbose around high level logic such as autoscaling/rest api.
* `3` - more verbose around lower level logic such as metric gathering and evaluation.
