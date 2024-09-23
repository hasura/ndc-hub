# Hasura DuckDB Connector

<a href="https://duckdb.org/"><img src="https://github.com/hasura/ndc-duckdb/blob/main/docs/logo.svg" align="right" width="200"></a>

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/connectors/duckdb)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-duckdb-blue.svg?style=flat)](https://hasura.io/connectors/duckdb)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://github.com/hasura/ndc-duckdb/blob/main/LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](https://github.com/hasura/ndc-duckdb/blob/main/README.md)

The Hasura DuckDB Connector allows for connecting to a DuckDB database or a MotherDuck hosted DuckDB database to give
you an instant GraphQL API on top of your DuckDB data.

This connector is built using the [Typescript Data Connector SDK](https://github.com/hasura/ndc-sdk-typescript) and
implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [See the listing in the Hasura Hub](https://hasura.io/connectors/duckdb)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0/index/)

## Features

Below, you'll find a matrix of all supported features for the DuckDB connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models | ❌        |       |
| Simple Object Query             | ✅        |       |
| Filter / Search                 | ✅        |       |
| Simple Aggregation              | ❌        |       |
| Sort                            | ✅        |       |
| Paginate                        | ✅        |       |
| Table Relationships             | ✅        |       |
| Views                           | ❌        |       |
| Distinct                        | ❌        |       |
| Remote Relationships            | ✅        |       |
| Custom Fields                   | ❌        |       |
| Mutations                       | ❌        |       |

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)
5. Have a [MotherDuck](https://motherduck.com/) hosted DuckDB database, or a persitent DuckDB database file — for
   supplying data to your API.

The steps below explain how to initialize and configure a connector for local development. You can learn how to deploy a
connector — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the DuckDB connector

With the [context set](https://hasura.io/docs/3.0/cli/commands/ddn_context_set/) for an existing subgraph, initialize
the connector:

```sh
ddn connector init -i
```

When the wizard runs, you'll be prompted to enter the following env vars necessary for your connector to function:

| Name       | Description                                                                                 |
| ---------- | ------------------------------------------------------------------------------------------- |
| DUCKDB_URL | The connection string for the DuckDB database, or the file path to the DuckDB database file |

If you are attaching to a local DuckDB file, first make sure that the file is located inside the connector directory.
For example, if you had a `data.duckdb` file you could place it at `/app/connector/duckdb/data.duckdb`. Files in the
connector directory get mounted to `/etc/connector/`.

**Your experience mounting files may vary, and while useful to explore a file locally, it's not recommended to attempt
to deploy a connector using a locally mounted file.**

After the CLI initializes the connector, you'll need to:

- [Introspect](https://hasura.io/docs/3.0/cli/commands/ddn_connector_introspect) the source.
- Add your [models](https://hasura.io/docs/3.0/cli/commands/ddn_model_add),
  [commands](https://hasura.io/docs/3.0/cli/commands/ddn_command_add), and
  [relationships](https://hasura.io/docs/3.0/cli/commands/ddn_relationship_add).
- Create a [new build](https://hasura.io/docs/3.0/cli/commands/ddn_supergraph_build_local).
- Test it by [running your project along with the connector](https://hasura.io/docs/3.0/cli/commands/ddn_run#examples).

## Documentation

View the full documentation for the DuckDB connector
[here](https://github.com/hasura/ndc-duckdb/blob/main/docs/index.md).

## Contributing

Check out our [contributing guide](https://github.com/hasura/ndc-duckdb/blob/main/docs/contributing.md) for more
details.

## License

The DuckDB connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
