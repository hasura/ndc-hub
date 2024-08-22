# Hasura DuckDB Connector
<a href="https://duckdb.org/"><img src="https://github.com/hasura/ndc-duckdb/blob/main/docs/logo.svg" align="right" width="200"></a>

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/connectors/duckdb)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-duckdb-blue.svg?style=flat)](https://hasura.io/connectors/duckdb)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://github.com/hasura/ndc-duckdb/blob/main/LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](https://github.com/hasura/ndc-duckdb/blob/main/README.md)

The Hasura DuckDB Connector allows for connecting to a DuckDB database or a MotherDuck hosted DuckDB database to give you an instant GraphQL API on top of your DuckDB data.

This connector is built using the [Typescript Data Connector SDK](https://github.com/hasura/ndc-sdk-typescript) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [See the listing in the Hasura Hub](https://hasura.io/connectors/duckdb)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0/index/)

## Features

Below, you'll find a matrix of all supported features for the DuckDB connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models | ❌     |       |
| Simple Object Query             | ✅     |       |
| Filter / Search                 | ✅     |       |
| Simple Aggregation              | ❌     |       |
| Sort                            | ✅     |       |
| Paginate                        | ✅     |       |
| Table Relationships             | ✅     |       |
| Views                           | ❌     |       |
| Distinct                        | ❌     |       |
| Remote Relationships            | ✅     |       |
| Custom Fields                   | ❌     |       |
| Mutations                       | ❌     |       |

## Before you get Started

1. The [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and [Docker](https://docs.docker.com/engine/install/) installed
2. A [supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
3. A [subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)
4. Have a [MotherDuck](https://motherduck.com/) hosted DuckDB database, or a persitent DuckDB database file — for supplying data to your API.

The steps below explain how to Initialize and configure a connector for local development. You can learn how to deploy a
connector — after it's been configured — [here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the DuckDB connector

### Step 1: Authenticate your CLI session

```bash
ddn auth login
```

### Step 2: Configure the connector

Once you have an initialized supergraph and subgraph, run the initialization command in interactive mode while providing a name for the connector in the prompt:

```bash
ddn connector init duckdb -i
```

#### Step 2.1: Choose the `hasura/duckdb` option from the list

#### Step 2.2: Choose a port for the connector

The CLI will ask for a specific port to run the connector on. Choose a port that is not already in use or use the default suggested port.

#### Step 2.3: Provide the env var(s) for the connector

| Name | Description |
|-|-|
| DUCKDB_URL       | The connection string for the DuckDB database, or the file path to the DuckDB database file |

You'll find the environment variables in the `.env` file and they will be in the format:

`<SUBGRAPH_NAME>_<CONNECTOR_NAME>_<VARIABLE_NAME>`

Here is an example of what your `.env` file might look like:

```
APP_DUCKDB_AUTHORIZATION_HEADER="Bearer SPHZWfL7P3Jdc9mDMF9ZNA=="
APP_DUCKDB_DUCKDB_URL="md:?motherduck_token=ey..."
APP_DUCKDB_HASURA_SERVICE_TOKEN_SECRET="SPHZWfL7P3Jdc9mDMF9ZNA=="
APP_DUCKDB_OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="http://local.hasura.dev:4317"
APP_DUCKDB_OTEL_SERVICE_NAME="app_duckdb"
APP_DUCKDB_READ_URL="http://local.hasura.dev:7525"
APP_DUCKDB_WRITE_URL="http://local.hasura.dev:7525"
```

### Step 3: Introspect the connector

Introspecting the connector will generate a `config.json` file and a `duckdb.hml` file.

```bash
ddn connector introspect duckdb
```

### Step 4: Add your resources

You can add the models, commands, and relationships to your API by tracking them which generates the HML files. 

```bash
ddn connector-link add-resources duckdb
```

## Documentation

View the full documentation for the DuckDB connector [here](https://github.com/hasura/ndc-duckdb/blob/main/docs/index.md).

## Contributing

Check out our [contributing guide](https://github.com/hasura/ndc-duckdb/blob/main/docs/contributing.md) for more details.

## License

The DuckDB connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).