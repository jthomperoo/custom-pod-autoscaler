# Rest API

Custom Pod Autoscalers expose a REST API that allows for interaction with the Custom Pod Autoscaler outside of configuration, at runtime.  

## Accessing the API

The API exposes the API at the [port specified in your configuration file (default `5000`)](../reference/configuration.md#apiconfig), and with the URL prefix `/api/<API_VERSION>`.  

Within a container the API can be accessed at `http://localhost:<PORT>/api/<API_VERSION>`.

## Configuring the API

See [the `apiConfig` section of the configuration reference](../reference/configuration.md#apiconfig).

## API Versions

The API is versioned with major versions, starting with `v1`. The API version does not neccessarily align with the major versions of the Custom Pod Autoscaler, which is versioned with Semantic Versioning.

### Versions

* [`v1`- Current, latest version](../reference/rest-api/v1.md).