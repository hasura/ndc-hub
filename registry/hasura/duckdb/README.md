# Hasura DuckDB Connector

<a href="https://duckdb.org/"><img src="https://github.com/hasura/ndc-duckdb/blob/main/docs/logo.svg" align="right" width="200"></a>

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/connectors/duckdb)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-duckdb-blue.svg?style=flat)](https://hasura.io/connectors/duckdb)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://github.com/hasura/ndc-duckdb/blob/main/LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](https://github.com/hasura/ndc-duckdb/blob/main/README.md)

The Hasura DuckDB Connector allows for connecting to a DuckDB database or a MotherDuck-hosted DuckDB database to give you an instant GraphQL API on top of your DuckDB data.

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

## Build on Hasura DDN

[Get started](https://hasura.io/docs/3.0/how-to-build-with-ddn/with-duckdb) by connecting your DuckDB instance to a Hasura DDN project.

## Fork the connector

You can fork the [connector's repo](https://github.com/hasura/ndc-duckdb) and iterate on it yourself.

## License

The DuckDB connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).

