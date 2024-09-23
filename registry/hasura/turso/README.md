# Hasura Turso Connector

<a href="https://turso.tech/"><img src="https://github.com/hasura/ndc-turso/blob/main/docs/logo.svg" align="right" width="200"></a>

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/connectors/turso)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-turso-blue.svg?style=flat)](https://hasura.io/connectors/turso)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://github.com/hasura/ndc-turso/blob/main/LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](https://github.com/hasura/ndc-turso/blob/main/README.md)

The Hasura Turso Connector allows for connecting to a LibSQL/SQLite database or a Turso hosted LibSQL database to give
you an instant GraphQL API on top of your Turso data.

This connector is built using the [Typescript Data Connector SDK](https://github.com/hasura/ndc-sdk-typescript) and
implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [See the listing in the Hasura Hub](https://hasura.io/connectors/turso)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0/index/)

## Features

Below, you'll find a matrix of all supported features for the Turso connector:

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
| Mutations                       | ✅        |       |

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

The steps below explain how to Initialize and configure a connector for local development. You can learn how to deploy a
connector — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the Turso connector

With the [context set](https://hasura.io/docs/3.0/cli/commands/ddn_context_set/) for an existing subgraph, initialize
the connector:

```sh
ddn connector init -i
```

When the wizard runs, you'll be prompted to enter the following env vars necessary for your connector to function:

| Name             | Description                                                                       |
| ---------------- | --------------------------------------------------------------------------------- |
| TURSO_URL        | The connection string for the Turso database, or the file path to the SQLite file |
| TURSO_AUTH_TOKEN | The turso auth token                                                              |

If you are attaching to a local SQLite file, first make sure that the file is located inside the connector directory.
For example, if you had a `data.sqlite` file you could place it at `/app/connector/turso/data.sqlite`. Files in the
connector directory get mounted to `/etc/connector/`.

In this instance, you would set the `TURSO_URL=file:/etc/connector/data.sqlite` and leave the `TURSO_AUTH_TOKEN` as
blank/null.

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

View the full documentation for the Turso connector [here](https://github.com/hasura/ndc-turso/blob/main/docs/index.md).

## Contributing

Check out our [contributing guide](https://github.com/hasura/ndc-turso/blob/main/docs/contributing.md) for more details.

## License

The Turso connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
