# Trino Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/getting-started/overview/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-snowflake-blue.svg?style=flat)](https://hasura.io/connectors/trino)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](./readme.md)

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your data models in
Trino. This connector supports Trino's functionalities listed in the table below, allowing for efficient and
scalable data operations. Additionally, users benefit from all the powerful features of Hasura’s Data Delivery Network
(DDN) platform, including query pushdown capabilities that delegate query operations to the database, thereby enhancing
query optimization and performance.

## Features

Below, you'll find a matrix of all supported features for the Trino connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models | ✅        |       |
| Native Mutations                | ❌        |       |
| Simple Object Query             | ✅        |       |
| Filter / Search                 | ✅        |       |
| Simple Aggregation              | ✅        |       |
| Sort                            | ✅        |       |
| Paginate                        | ✅        |       |
| Table Relationships             | ✅        |       |
| Views                           | ✅        |       |
| Remote Relationships            | ✅        |       |
| Custom Fields                   | ❌        |       |
| Mutations                       | ❌        |       |
| Distinct                        | ❌        |       |
| Enums                           | ❌        |       |
| Naming Conventions              | ❌        |       |
| Default Values                  | ❌        |       |
| User-defined Functions          | ❌        |       |

## Build on Hasura DDN

[Get started](https://hasura.io/docs/3.0/how-to-build-with-ddn/with-trino) by connecting your Trino query engine to a Hasura DDN project.

## Fork the connector

You can fork the [connector's repo](https://github.com/hasura/trino-data-connector) and iterate on it yourself.

## License

The Hasura Snowflake connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
