## BigQuery Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/getting-started/overview/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-bigquery-blue.svg?style=flat)](https://hasura.io/connectors/bigquery-jdbc)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](./readme.md)

The Hasura BigQuery connector allows you to connect your Hasura DDN instance to Google BigQuery for efficient, scalable
data operations.

## Features

Below, you'll find a matrix of all supported features for the BigQuery connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models | ❌        |       |
| Native Mutations                | ❌        |       |
| Simple Object Query             | ✅        |       |
| Filter / Search                 | ✅        |       |
| Simple Aggregation              | ✅        |       |
| Sort                            | ✅        |       |
| Paginate                        | ✅        |       |
| Table Relationships             | ❌        |       |
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

[Get started](https://hasura.io/docs/3.0/how-to-build-with-ddn/with-bigquery) by connecting BigQUery to a Hasura DDN project.

## Fork the connector

You can fork the [connector's repo](https://github.com/hasura/ndc-jdbc) and iterate on it yourself.

## License

The Hasura BigQuery connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
