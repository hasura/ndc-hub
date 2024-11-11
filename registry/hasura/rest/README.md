## Overview

The Hasura REST Connector allows for connecting Hasura to REST API servers giving you an instant GraphQL API on top of your REST data.

Data Connectors are the way to connect the Hasura Data Delivery Network (DDN) to external data sources. A data connector is an HTTP service that exposes a set of APIs that Hasura uses to communicate with the data source. Data connectors are built to conform to the [NDC Specification](https://hasura.github.io/ndc-spec/overview.html) using one of Hasura's available SDKs. The data connector is responsible for interpreting work to be done on behalf of the Hasura Engine, using the native query language of the data source.

The data connector is open source and can be found in the [ndc-rest GitHub repository](https://github.com/hasura/ndc-rest).

## Features

- No code, configuration based.
- Supported many API specifications.
- Composable API connections.
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

- `application/json`
- `application/x-www-form-urlencoded`
- `application/octet-stream`
- `multipart/form-data`
- `text/*`
- Upload file content types, e.g.`image/*` from `base64` arguments.

## Deployment

The connector is hosted by Hasura and can be used from the [Hasura v3 Console](https://console.hasura.io).

## Usage

The Hasura REST connector can be deployed using the [Hasura CLI](https://hasura.io/docs/3.0/cli/overview) by following either the [Quick Start Guide](https://hasura.io/docs/3.0/getting-started/overview/) or [deploying the connector](https://hasura.io/docs/3.0/connectors/deployment).

## Troubleshooting

Please [submit a Github issue](https://github.com/hasura/ndc-rest/issues/new)
if you encounter any problems!
