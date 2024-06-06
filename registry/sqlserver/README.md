## Overview

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/getting-started/overview/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-sqlserver-blue.svg?style=flat)](https://hasura.io/connectors/sqlserver)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](./readme.md)

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your data models in
Microsoft SQL Server. This connector supports SQL Server's functionalities listed in the table below, allowing for
efficient and scalable data operations. Additionally, users benefit from all the powerful features of Hasura’s Data
Delivery Network (DDN) platform, including query pushdown capabilities that delegate query operations to the database,
thereby enhancing query optimization and performance.

This connector is built using the [Rust Data Connector SDK](https://github.com/hasura/ndc-hub#rusk-sdk) and implements
the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [Connector information in the Hasura Hub](https://hasura.io/connectors/sqlserver)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0)

## Features

Below, you'll find a matrix of all supported features for the SQL Server connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models |    ✅     |       |
| Simple Object Query             |    ✅     |       |
| Filter / Search                 |    ✅     |       |
| Simple Aggregation              |    ✅     |       |
| Sort                            |    ✅     |       |
| Paginate                        |    ✅     |       |
| Table Relationships             |    ✅     |       |
| Views                           |    ✅     |       |
| Remote Relationships            |    ✅     |       |
| Custom Fields                   |    ❌     |       |
| Mutations                       |    ❌     |       |
| Distinct                        |    ✅     |       |
| Enums                           |    ❌     |       |
| Naming Conventions              |    ❌     |       |
| Default Values                  |    ❌     |       |
| User-defined Functions          |    ❌     |       |

## Before you get Started

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Install the [CLI](https://hasura.io/docs/3.0/cli/installation/)
3. Install the [Hasura VS Code extension](https://marketplace.visualstudio.com/items?itemName=HasuraHQ.hasura)
4. [Create a project](https://hasura.io/docs/3.0/getting-started/create-a-project)

## Using the connector

To use the SQL Server connector, follow these steps in a Hasura project:

1. Add the connector:

   ```bash
   ddn add connector-manifest sqlserver_connector --subgraph app --hub-connector hasura/sqlserver --type cloud
   ```

   In the snippet above, we've used the subgraph `app` as it's available by default; however, you can change this value
   to match any [subgraph](https://hasura.io/docs/3.0/project-configuration/subgraphs) which you've created in your
   project.

2. Add your connection URI:

   Open your project in your text editor and open the `/app/sqlserver_connector/connector/sqlserver_connector.build.hml`
   file of your project. Then, add the `CONNECTION_URI` environment variable with the connection string:

   ```yaml
   # other configuration above
   CONNECTION_URI:
     value: "<YOUR_CONNECTION_URI>"
   ```

3. Update the connector manifest and the connector link

   These two steps will (1) allow Hasura to introspect your data source and complete the configuration and (2) deploy
   the connector to Hasura DDN:

   ```bash
   ddn update connector-manifest sqlserver_connector
   ```

   ```bash
   ddn update data-connector-link sqlserver_connector --add-all-resources
   ```

   `--add-all-resources` flag adds all the models and commands present in the database to the connector metadata.

4. Create a build

   ```bash
   ddn build supergraph-manifest
   ```

   This will return information about the build:

   |               |                                                                                                   |
   | ------------- | ------------------------------------------------------------------------------------------------- |
   | Build Version | bd96bb221a                                                                                        |
   | API URL       | https://<PROJECT_NAME>-bd96bb221a.ddn.hasura.app/graphql                                          |
   | Console URL   | https://console.hasura.io/project/<PROJECT_NAME>/environment/default/build/bd96bb221a/graphql     |
   | Project Name  | <PROJECT_NAME>                                                                                    |
   | Description   |                                                                                                   |

   Follow the project configuration build [guide](https://hasura.io/docs/3.0/project-configuration/builds/) to apply
   other actions on the build.

5. Test the API

   The console URL in the build information cna be used to open the GraphiQL console to test out the API

## Contributing

We're happy to receive any contributions from the community. Please refer to our [development guide](https://github.com/hasura/ndc-sqlserver/blob/main/docs/development.md).

## License

The Hasura SQL Server connector is available under the [Apache License
2.0](https://www.apache.org/licenses/LICENSE-2.0).
