# Hasura Turso Connector
<a href="https://turso.tech/"><img src="https://github.com/hasura/ndc-turso/blob/main/docs/logo.svg" align="right" width="200"></a>


[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/connectors/turso)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-turso-blue.svg?style=flat)](https://hasura.io/connectors/turso)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://github.com/hasura/ndc-turso/blob/main/LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](https://github.com/hasura/ndc-turso/blob/main/README.md)

The Hasura Turso Connector allows for connecting to a LibSQL/SQLite database or a Turso hosted LibSQL database to give you an instant GraphQL API on top of your Turso data.

This connector is built using the [Typescript Data Connector SDK](https://github.com/hasura/ndc-sdk-typescript) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [See the listing in the Hasura Hub](https://hasura.io/connectors/turso)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0/index/)

## Features

Below, you'll find a matrix of all supported features for the Turso connector:

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
| Mutations                       | ✅     |       |

## Before you get Started

1. The [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and [Docker](https://docs.docker.com/engine/install/) installed
2. A [supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
3. A [subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)
4. Have a [Turso](https://turso.tech/) hosted database, or a persistent Turso SQLite database file — for supplying data to your API.

The steps below explain how to Initialize and configure a connector for local development. You can learn how to deploy a
connector — after it's been configured — [here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the Turso connector

### Step 1: Authenticate your CLI session

```bash
ddn auth login
```

### Step 2: Configure the connector

Once you have an initialized supergraph and subgraph, run the initialization command in interactive mode while providing a name for the connector in the prompt:

```bash
ddn connector init turso -i
```

#### Step 2.1: Choose the `hasura/turso` option from the list

#### Step 2.2: Choose a port for the connector

The CLI will ask for a specific port to run the connector on. Choose a port that is not already in use or use the default suggested port.

#### Step 2.3: Provide the env var(s) for the connector

| Name | Description |
|-|-|
| TURSO_URL        | The connection string for the Turso database, or the file path to the SQLite file |
| TURSO_AUTH_TOKEN | The turso auth token |

You'll find the environment variables in the `.env` file and they will be in the format:

`<SUBGRAPH_NAME>_<CONNECTOR_NAME>_<VARIABLE_NAME>`

Here is an example of what your `.env` file might look like:

```
APP_TURSO_AUTHORIZATION_HEADER="Bearer QTJ7rl19SvKa0rwOZjYILQ=="
APP_TURSO_HASURA_SERVICE_TOKEN_SECRET="QTJ7rl19SvKa0rwOZjYILQ=="
APP_TURSO_OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="http://local.hasura.dev:4317"
APP_TURSO_OTEL_SERVICE_NAME="app_turso"
APP_TURSO_READ_URL="http://local.hasura.dev:4362"
APP_TURSO_TURSO_AUTH_TOKEN="eyJ..."
APP_TURSO_TURSO_URL="libsql://chinook-tristenharr.turso.io"
APP_TURSO_WRITE_URL="http://local.hasura.dev:4362"
```

If you are attaching to a local SQLite file, first make sure that the file is located inside the connector directory. For example, if you had a `data.sqlite` file you could place it at `/app/connector/turso/data.sqlite`. Files in the connector directory get mounted to `/etc/connector/`. 

In this instance, you would set the `TURSO_URL=/etc/connector/data.sqlite` and leave the `TURSO_AUTH_TOKEN` as blank/null. Now your `.env` might look like this:

```
APP_TURSO_AUTHORIZATION_HEADER="Bearer QTJ7rl19SvKa0rwOZjYILQ=="
APP_TURSO_HASURA_SERVICE_TOKEN_SECRET="QTJ7rl19SvKa0rwOZjYILQ=="
APP_TURSO_OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="http://local.hasura.dev:4317"
APP_TURSO_OTEL_SERVICE_NAME="app_turso"
APP_TURSO_READ_URL="http://local.hasura.dev:4362"
APP_TURSO_TURSO_URL="/etc/connector/data.sqlite"
APP_TURSO_WRITE_URL="http://local.hasura.dev:4362"
```

Your experience mounting files may vary, and while useful to explore a file locally, it's not recommended to attempt to deploy a connector using a locally mounted file.

### Step 3: Introspect the connector

Introspecting the connector will generate a `config.json` file and a `turso.hml` file.

```bash
ddn connector introspect turso
```

### Step 4: Add your resources

You can add the models, commands, and relationships to your API by tracking them which generates the HML files. 

```bash
ddn connector-link add-resources turso
```

## Documentation

View the full documentation for the Turso connector [here](https://github.com/hasura/ndc-turso/blob/main/docs/index.md).

## Contributing

Check out our [contributing guide](https://github.com/hasura/ndc-turso/blob/main/docs/contributing.md) for more details.

## License

The Turso connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).