# Hasura Turso Connector

<a href="https://turso.tech/"><img src="https://github.com/hasura/ndc-turso/blob/main/docs/logo.svg" align="right" width="200"></a>

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/connectors/turso)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-turso-blue.svg?style=flat)](https://hasura.io/connectors/turso)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://github.com/hasura/ndc-turso/blob/main/LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](https://github.com/hasura/ndc-turso/blob/main/README.md)

The Hasura Turso Connector allows for connecting to a LibSQL/SQLite database or a Turso hosted LibSQL database to give
you an instant GraphQL API on top of your Turso data.

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

## Build on Hasura DDN

[Get started](https://hasura.io/docs/3.0/how-to-build-with-ddn/with-turso) by connecting your Turso instance to a Hasura DDN project.

## Fork the connector

You can fork the [connector's repo](https://github.com/hasura/ndc-turso) and iterate on it yourself.

## Contributing

Check out our [contributing guide](https://github.com/hasura/ndc-turso/blob/main/docs/contributing.md) for more details.

## License

The Turso connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
