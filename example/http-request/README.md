# HTTP Request

This is a simple autoscaler designed to show how to use a HTTP method as part of
a Custom Pod Autoscaler. This example reaches out to the
[random.org](http://random.org/) HTTP API to generate a random integer between 1
and 5 and uses that to decide what value to scale the managed resource to.

## Overview

The metric gathering stage of this autoscaler uses the HTTP method to reach out
to [https://www.random.org/integers/](https://www.random.org/integers/) to
generate a random integer. This HTTP method is defined in `config.yaml`:

```yaml
metric:
  type: "http"
  timeout: 2500
  http:
    method: "GET"
    url: "https://www.random.org/integers/?num=1&min=1&max=5&col=1&base=10&format=plain&rnd=new"
    successCodes:
      - 200
    parameterMode: query
```

This makes a `GET` request to
`https://www.random.org/integers/?num=1&min=1&max=5&col=1&base=10&format=plain&rnd=new`.

This URL has some parameters specific to the `random.org` API:
- `num=1` means return only 1 number.
- `min=1` means minimum value 1.
- `max=5` means maximum value 5.
- `col=1` means return the results in a single column.
- `base=10` means numbers in base 10.

The HTTP request has a timeout of `2500` milliseconds (2.5 seconds), and is only
presumed successful if it recieves a response code of `200`. Additional
information is provided by query parameter, but for this example it is
unimportant and is ignored.

The autoscaler then takes the results of this and feeds it into `evaluator.py`
which wraps it in JSON and returns it back to the autoscaler to scale the
resource being managed.
