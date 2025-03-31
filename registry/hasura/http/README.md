## Overview

The Hasura HTTP Connector allows for connecting Hasura to HTTP API servers giving you an instant GraphQL API on top of your RESTful data.

Data Connectors are the way to connect the Hasura Data Delivery Network (DDN) to external data sources. A data connector is an HTTP service that exposes a set of APIs that Hasura uses to communicate with the data source. Data connectors are built to conform to the [NDC Specification](https://hasura.github.io/ndc-spec/overview.html) using one of Hasura's available SDKs. The data connector is responsible for interpreting work to be done on behalf of the Hasura Engine, using the native query language of the data source.

The data connector is open source and can be found in the [ndc-http GitHub repository](https://github.com/hasura/ndc-http).

> [!NOTE]
> HTTP connector is a configuration-based HTTP engine and isn't limited to the OpenAPI specs only. Use [OpenAPI Connector](https://hasura.io/docs/3.0/connectors/external-apis/open-api) if you want to take more control of OpenAPI via code generation.

## Features

- No code, configuration-based.
- Supported many API specifications.
- Composable API collections.
- Supported authentication.
- Supported headers forwarding.
- Supported timeout and retry.
- Supported concurrency and sending distributed requests to multiple servers.

**Supported request types**

| Request Type | Query | Path | Body | Headers |
| ------------ | ----- | ---- | ---- | ------- |
| GET          | ✅    | ✅   | NA   | ✅      |
| POST         | ✅    | ✅   | ✅   | ✅      |
| DELETE       | ✅    | ✅   | ✅   | ✅      |
| PUT          | ✅    | ✅   | ✅   | ✅      |
| PATCH        | ✅    | ✅   | ✅   | ✅      |

**Supported content types**

| Content Type                      | Supported |
| --------------------------------- | --------- |
| application/json                  | ✅        |
| application/xml                   | ✅        |
| application/x-www-form-urlencoded | ✅        |
| multipart/form-data               | ✅        |
| text/\*                           | ✅        |
| application/x-ndjson              | ✅        |
| application/octet-stream          | ✅ (\*)   |
| image/\*                          | ✅ (\*)   |
| video/\*                          | ✅ (\*)   |

(\*) Upload file content types are converted to `base64` encoding.

**Supported authentication**

| Security scheme | Supported | Comment                                                                                                                                   |
| --------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| API Key         | ✅        |                                                                                                                                           |
| Bearer Auth     | ✅        |                                                                                                                                           |
| Cookies         | ✅        | Require forwarding the `Cookie` header from the Hasura engine.                                                                            |
| OAuth 2.0       | ✅        | Built-in support for the `client_credentials` grant. Other grant types require forwarding access tokens from headers by the Hasura engine |
| mTLS            | ✅        |                                                                                                                                           |

## Build on Hasura DDN

[Get started](https://hasura.io/docs/3.0/how-to-build-with-ddn/with-http) by connecting your existing API to a Hasura DDN project.

## Fork the connector

You can fork the [connector's repo](https://github.com/hasura/ndc-http) and iterate on it yourself.

## Troubleshooting

Please [submit a Github issue](https://github.com/hasura/ndc-http/issues/new)
if you encounter any problems!
