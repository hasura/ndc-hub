# Hasura Qdrant Connector
<a href="https://qdrant.tech/"><img src="https://github.com/hasura/ndc-qdrant/blob/main/docs/logo.png" align="right" width="200"></a>

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/connectors/qdrant)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-qdrant-blue.svg?style=flat)](https://hasura.io/connectors/qdrant)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://github.com/hasura/ndc-qdrant/blob/main/LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](https://github.com/hasura/ndc-qdrant/blob/main/README.md)

The Hasura Qdrant Connector allows for connecting to a Qdrant database to give you an instant GraphQL API on top of your Qdrant data.

This connector is built using the [Typescript Data Connector SDK](https://github.com/hasura/ndc-sdk-typescript) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [See the listing in the Hasura Hub](https://hasura.io/connectors/qdrant)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0/index/)

## Features

Below, you'll find a matrix of all supported features for the Qdrant connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models | ❌     |       |
| Simple Object Query             | ✅     |       |
| Filter / Search                 | ✅     |       |
| Simple Aggregation              | ❌     |       |
| Sort                            | ❌     |       |
| Paginate                        | ✅     | Pagination offset field only works for documents with Integer ID's       |
| Nested Objects                  | ✅     |       |
| Nested Arrays                   | ✅     |       |
| Nested Filtering                | ❌     |       |
| Nested Sorting                  | ❌     |       |
| Nested Relationships            | ❌     |       |
| Vector Search                   | ✅     |       |

## Before you get Started

[Prerequisites or recommended steps before using the connector.]

1. The [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and [Docker](https://docs.docker.com/engine/install/) installed
2. A [supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
3. A [subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)
4. Have a [Qdrant](https://qdrant.tech/) hosted database, or a locally running Qdrant database — for supplying data to your API.

The steps below explain how to Initialize and configure a connector for local development. You can learn how to deploy a
connector — after it's been configured — [here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the Qdrant connector

### Step 1: Authenticate your CLI session

```bash
ddn auth login
```

### Step 2: Configure the connector

Once you have an initialized supergraph and subgraph, run the initialization command in interactive mode while providing a name for the connector in the prompt:

```bash
ddn connector init qdrant -i
```

#### Step 2.1: Choose the `hasura/qdrant` option from the list

#### Step 2.2: Choose a port for the connector

The CLI will ask for a specific port to run the connector on. Choose a port that is not already in use or use the default suggested port.

#### Step 2.3: Provide the env var(s) for the connector

| Name | Description |
|-|-|
| QDRANT_URL        | The connection string for the Qdrant database |
| QDRANT_API_KEY | The Qdrant API Key |

You'll find the environment variables in the `.env` file and they will be in the format:

`<SUBGRAPH_NAME>_<CONNECTOR_NAME>_<VARIABLE_NAME>`

Here is an example of what your `.env` file might look like:

```
APP_QDRANT_AUTHORIZATION_HEADER="Bearer B9-PceSL1QrUE_Z1gJNdGQ=="
APP_QDRANT_HASURA_SERVICE_TOKEN_SECRET="B9-PceSL1QrUE_Z1gJNdGQ=="
APP_QDRANT_OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="http://local.hasura.dev:4317"
APP_QDRANT_OTEL_SERVICE_NAME="app_qdrant"
APP_QDRANT_QDRANT_API_KEY="5PX..."
APP_QDRANT_QDRANT_URL="https://2a4ae326-fdef-473e-a13c-7dc07f2f2759.us-east4-0.gcp.cloud.qdrant.io"
APP_QDRANT_READ_URL="http://local.hasura.dev:5963"
APP_QDRANT_WRITE_URL="http://local.hasura.dev:5963"
```

### Step 3: Introspect the connector

Introspecting the connector will generate a `config.json` file and a `qdrant.hml` file.

```bash
ddn connector introspect qdrant
```

### Step 4: Add your resources

You can add the models, commands, and relationships to your API by tracking them which generates the HML files. 

```bash
ddn connector-link add-resources qdrant
```

## Documentation

View the full documentation for the Qdrant connector [here](https://github.com/hasura/ndc-qdrant/blob/main/docs/index.md).

## Contributing

Check out our [contributing guide](https://github.com/hasura/ndc-qdrant/blob/main/docs/contributing.md) for more details.

## License

The Qdrant connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
