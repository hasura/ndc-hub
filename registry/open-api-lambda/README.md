# OpenAPI Lambda Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/getting-started/overview/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-postgres-blue.svg?style=flat)](https://hasura.io/connectors)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)

The OpenAPI Lambda Connector allows you to import APIs that are documented in the OpenAPI/Swagger format into the Hasura Supergraph. The connector exposes REST API endpoints as Typescript functions, which can be exposed as GraphQL queries or mutations via the [NodeJS Lambda Connector](https://github.com/hasura/ndc-nodejs-lambda).

Functions that wrap GET requests are marked with a `@readonly` annotation, and are exposed as GraphQL Queries by the [NodeJS Lambda Connector](https://github.com/hasura/ndc-nodejs-lambda). All other request types are exposed as GraphQL Mutations.

This Connector implements the [Data Connector Spec](https://github.com/hasura/ndc-spec)

- [Hasura V3 Documentation](https://hasura.io/docs/3.0)
- [NodeJS Lambda Connector](https://github.com/hasura/ndc-nodejs-lambda)

## Features

- Convert Open API/swagger documentation into Typescript functions compatible with NodeJS Lambda Connector
- Supported request types

| Request Type | Query | Path | Body | Headers |
| ------------ | ----- | ---- | ---- | ------- |
| GET          | ✅    | ✅   | NA   | ✅      |
| POST         | ✅    | ✅   | ✅   | ✅      |
| DELETE       | ✅    | ✅   | ✅   | ✅      |
| PUT          | ✅    | ✅   | ✅   | ✅      |
| PATCH        | ✅    | ✅   | ✅   | ✅      |

## Before you get Started

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Install the [DDN CLI](https://hasura.io/docs/3.0/cli/installation/)
3. Please ensure that you have Docker installed and the Docker Daemon is running.
4. If you want to make changes to the generated Typescript files, please ensure you have Node.js v20+ installed

## Quickstart using the DDN CLI

> [!TIP]
> The following instructions are just a quick summary of how to use the OpenAPI Lambda connector.
> To see it in use in a wider Hasura DDN project, and to understand the underlying DDN concepts, please check out the [Hasura DDN Getting Started Guide](https://hasura.io/docs/3.0/getting-started/overview/).

> [!NOTE]
> This section assumes that you have already setup a Supergraph and added a Subgraph.

1. Initialize the connector

```
ddn connector init my_openapi --subgraph my_subgraph --hub-connector hasura/openapi
```

This will generate the necessary files into `my_subgraph/connector/my_openapi` directory. Supporting Typescript files for API calls will be created in this directory.

2. Add the correct environment variables to `/my_subgraph/connector/my_openapi/.env.local`. Supported environment variables and their description are listed under [Supported Environment Variables](#supported-environment-variables) section.

3. Modify the Docker container port in `my_subgraph/connector/my_openapi/docker-compose.my_openapi.yaml`. Typically, connectors default to port 8080. Each time you add a connector, please increment the published port by one to avoid port collisions. For example:

```
ports:
  - mode: ingress
    target: 8080
    published: '8082'
    protocol: tcp
```

4. Introspect the OpenAPI document using the connector

```
ddn connector introspect --connector my_subgraph/connector/my_openapi/connector.yaml
```

This will introspect your OpenAPI document and create an `api.ts` file, a `functions.ts` file and other supporting files required to run the Typescript project.

- The `api.ts` file contains the Data Types and API calls from the OpenAPI document
- The `functions.ts` file contains functions that wrap API calls. You can modify this `functions.ts` file to introduce business logic if you want to. See [Saving User Changes](#saving-user-changes) if you want to preserve your changes in this file when you introspect the OpenAPI document again.

5. Add a [Data Connector Link](https://hasura.io/docs/3.0/supergraph-modeling/data-connector-links)

```
ddn connector-link add my_openapi --subgraph my_subgraph
```

This will create a file `my_subgraph/metadata/my_openapi.hml` that links your OpenAPI Connector to your Hasura Supergraph.

6. Update the evironment variables listed in `my_subgraph/metadata/my_openapi.hml` (here, `MY_SUBGRAPH_MY_OPENAPI_READ_URL` and `MY_SUBGRAPH_MY_OPENAPI_WRITE_URL`) in `my_subgraph/.env.my_subgraph`.

7. Start the connector in a Docker container

```
docker compose -f my_subpgraph/connector/my_openapi/docker-compose.my_openapi.yaml up
```

The `http://localhost:${your-docker-container-port}/schema/` should return the schema of your OpenAPI Connector. All functions that wrap `GET` requests will be listed in the `functions` array, while all other functions will be listed in the `procedures` array.

8. Add the connector schema to [Data Connector Link](<(https://hasura.io/docs/3.0/supergraph-modeling/data-connector-links)>)

```
ddn connector-link update my_openapi --subgraph my_subgraph
```

This command will modify `my_subgraph/metadata/my_openapi.hml`, and the schema of the connector will be added to the `definition.schema` key.

9. Add all resources in your connector schema to your GraphQL API

```
ddn connector-link update my_openapi --subgraph my_subgraph --add-all-resources
```

This will create HML files that represent resources in your connector schema at `my_subgraph/metadata`. These HML files specify your GraphQL API.

You have now added the OpenAPI Connector and imported all of your APIs in your supergraph. You can now:

- Deploy the supergraph to Hasura DDN (Please follow the steps [here](https://hasura.io/docs/3.0/getting-started/deployment/))
- Run the supergraph locally for debugging (Please follow the steps [here](https://hasura.io/docs/3.0/getting-started/build-your-api))

## Documentation

This connector is published as a Docker Image. The image name is `ghcr.io/hasura/ndc-open-api-lambda`. The Docker Image accepts the following environment variables that can be used to alter its functionality.

### Supported Environment Variables

1. `NDC_OAS_DOCUMENT_URI` (optional): The URI to your Open API Document. If you're using a file instead of a HTTP link, please ensure that it is named `swagger.json` and is present in the root directory of the volume being mounted to `/etc/connector` (for this tutorial, the `swagger.json` file should be present at `my_subgraph/connector/my_openapi/`).
2. `NDC_OAS_BASE_URL` (optional): The base URL of your API.
3. `NDC_OAS_FILE_OVERWRITE` (optional): Boolean flag to allow previously generated files to be over-written. Defaults to `false`.
4. `HASURA_PLUGIN_LOG_LEVEL` (optional): The log level. Possible values: `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`. Defaults to `info`
5. `NDC_OAS_LAMBDA_PRETTY_LOGS` (optional): Boolean flag to print human readable logs instead of JSON. Defaults to `false`

### Saving User Changes

When re-introspecting the connector, user changes in `functions.ts` can be preserved by adding an `@save` JS Doc Tag to the documentation comment of a function. This will ensure that that function is not overwritten and the saved function will be added if missing in the newly generated `functions.ts`

Example

```
/**
 * Dummy function that mutates an API response
 * @save
 */
function mutateResponse(response: ApiResponseObject) {
  response.description = "This API does some work. I hope that's helpful";
}
```
