# Methods

Methods specify how user logic should be called by the Custom Pod Autoscaler base program.

## shell

The shell method allows specifying a shell command, run through `/bin/sh`. Any relevant
information will be provided to the command specified by piping the information in through standard
in. Data is returned by writing to standard out. An error is signified by exiting with a non-zero
exit code; if an error occurs the autoscaler will capture all standard error and out and log it.

### Example

This is an example configuration of the shell method for the metric gatherer:
```yaml
metric:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "python"
    command:
      - "/metric.py"
    logStderr: false
```
Breaking this example down:

- `type` = the type of the method, for this example it is a `shell` method.
- `timeout` = the maximum time the method can take in milliseconds, for this example it is `2500` (2.5 seconds), if it takes longer than this it will count the method as failing.
- `shell` = the shell method to execute.
  - `entrypoint` = the entrypoint of the shell command, e.g. `/bin/bash`, defaults to `/bin/sh`.
  - `command` = the command to execute.
  - `logStderr` = an optional flag that logs a successful shell method's stderr, defaults to `false`.

### Always Fail Example

This is a metric configuration that will always fail:
```yaml
metric:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "/bin/sh"
    command:
      - "-c"
      - "exit 1"
```

### Always Return 5 Example

This is a metric configuration that will return `5` as a metric.
```yaml
metric:
  type: "shell"
  timeout: 2500
  shell:
    entrypoint: "/bin/sh"
    command:
      - "-c"
      - "echo '5'"
```

## http

The http method allows defining an HTTP request for the autoscaler to make. Any
relevant information will be provided to the target of the request by HTTP
parameters - either `query` or `body` parameters. An error is signified by a
status code that is not defined to be successful in the configuration; if this
kind of error occurs the autoscaler will capture the response body and log it.

### Example

This is an example configuration of the http method for the metric gatherer:

```yaml
metric:
  type: "http"
  timeout: 2500
  http:
    method: "GET"
    url: "https://www.custompodautoscaler.com"
    successCodes:
      - 200
    headers:
      exampleHeader: exampleHeaderValue
    parameterMode: query
    caCert: "/ca.pem"
    clientCert: "/client.cert"
    clientKey: "/client.key"
```

Breaking this example down:

- `type` = the type of the method, for this example it is an `http` method.
- `timeout` = the maximum time the method can take in milliseconds, for this
  example it is `2500` (2.5 seconds), if it takes longer than this it will count
  the method as failing.
- `http` = configuration of the HTTP request.
  - `method` = the HTTP method of the HTTP request.
  - `url` = the URL to target with the HTTP request.
  - `successCodes` = a list of success codes defining how to determine if the
    request is successful - if the request responds with a code not on this list
    it will be assumed to be a failure.
  - `headers` = a dictionary of headers that can be provided with the request,
    in this example the key is `exampleHeader` and the value is
    `exampleHeaderValue`. This is an optional parameter.
  - `parameterMode` = the mode for passing parameters to the target; either
    `query` - as a query parameter, or `body` - as a body parameter. In this
    example it is by query parameter.
  - `caCert` = an optional path to a CA certificate to trust for an HTTPS request.
  - `clientCert` = an optional client certificate to use for mutual TLS.
  - `clientKey` = an optional client key to use for mutual TLS.

### POST Example

This is an example using HTTP `POST` and information passed as a body parameter.

```yaml
metric:
  type: "http"
  timeout: 2500
  http:
    method: "POST"
    url: "https://www.custompodautoscaler.com"
    successCodes:
      - 200
      - 202
    parameterMode: body
```
