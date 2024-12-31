# SQL Server Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-sqlserver-blue.svg?style=flat)](https://hasura.io/connectors/sqlserver)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)

> **Note:** ADO.NET is the supported connection string format for SQL Server for ndc-sqlserver in DDN.
> You can find the documentation for ADO.NET SQL Server connection strings [here](https://learn.microsoft.com/en-us/dotnet/framework/data/adonet/connection-string-syntax#sqlclient-connection-strings).
> This is a change from Hasura version 2, where ODBC connection strings were supported.

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your data models in
Microsoft SQL Server. This connector supports SQL Server's functionalities listed in the table below, allowing for
efficient and scalable data operations. Additionally, users benefit from all the powerful features of Hasura’s Data
Delivery Network (DDN) platform, including query pushdown capabilities that delegate query operations to the database,
thereby enhancing query optimization and performance.

This connector is built using the [Rust Data Connector SDK](https://github.com/hasura/ndc-hub#rusk-sdk) and implements
the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [See the listing in the Hasura Hub](https://hasura.io/connectors/sqlserver)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0/)

## Features

Below, you'll find a matrix of all supported features for the SQL Server connector:

| Feature                         | Supported | Notes                                                                                                                                                    |
|---------------------------------|-----------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Native Queries + Logical Models | ✅        |                                                                                                                                                          |
| Native Mutations                | ✅        |                                                                                                                                                          |
| Simple Object Query             | ✅        |                                                                                                                                                          |
| Filter / Search                 | ✅        |                                                                                                                                                          |
| Simple Aggregation              | ✅        | The limit parameter does not work as expected when combined with aggregate functions. Currently, any limit value set in these cases will be disregarded. |
| Sort                            | ✅        |                                                                                                                                                          |
| Paginate                        | ✅        |                                                                                                                                                          |
| Table Relationships             | ✅        |                                                                                                                                                          |
| Views                           | ✅        |                                                                                                                                                          |
| Remote Relationships            | ✅        |                                                                                                                                                          |
| Stored Procedures               | ✅        |                                                                                                                                                          |
| Custom Fields                   | ❌        |                                                                                                                                                          |
| Mutations                       | ❌        | Only native mutations are suppported                                                                                                                     |
| Distinct                        | ✅        |                                                                                                                                                          |
| Enums                           | ❌        |                                                                                                                                                          |
| Naming Conventions              | ❌        |                                                                                                                                                          |
| Default Values                  | ❌        |                                                                                                                                                          |
| User-defined Functions          | ❌        |                                                                                                                                                          |

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/cli/commands/ddn_supergraph_init)

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes).You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the SQLServer connector

With the [context set](https://hasura.io/docs/3.0/cli/commands/ddn_context_set/) for an existing subgraph, initialize
the connector:

```sh
ddn connector init -i
```
> **Note:** The `CONNECTION_URI` is the connection string of the SQL Server database. You can find the documentation for ADO.NET SQL Server connection string formats [here](https://learn.microsoft.com/en-us/dotnet/framework/data/adonet/connection-string-syntax#sqlclient-connection-strings).

When the wizard runs, you'll be prompted to enter the following env vars necessary for your connector to function:

| Name           | Description                                      | Required | Default |
| -------------- | ------------------------------------------------ | -------- | ------- |
| CONNECTION_URI | The connection string of the SQL Server database | Yes      | N/A     |

After the CLI initializes the connector, you'll need to:

- [Introspect](https://hasura.io/docs/3.0/cli/commands/ddn_connector_introspect) the source.
- Add your [models](https://hasura.io/docs/3.0/cli/commands/ddn_model_add),
  [commands](https://hasura.io/docs/3.0/cli/commands/ddn_command_add), and
  [relationships](https://hasura.io/docs/3.0/cli/commands/ddn_relationship_add).
- Create a [new build](https://hasura.io/docs/3.0/cli/commands/ddn_supergraph_build_local).
- Test it by [running your project along with the connector](https://hasura.io/docs/3.0/cli/commands/ddn_run#examples).

## Documentation

View the full documentation for the ndc-sqlserver connector [here](https://github.com/hasura/ndc-sqlserver/blob/main/docs/readme.md).

## Contributing

We're happy to receive any contributions from the community. Please refer to our
[development guide](https://github.com/hasura/ndc-sqlserver/blob/main/docs/development.md).

## License

The Hasura SQL Server connector is available under the
[Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
