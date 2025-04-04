# Hasura Qdrant Connector

<a href="https://qdrant.tech/"><img src="https://github.com/hasura/ndc-qdrant/blob/main/docs/logo.png" align="right" width="200"></a>

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/connectors/qdrant)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-qdrant-blue.svg?style=flat)](https://hasura.io/connectors/qdrant)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://github.com/hasura/ndc-qdrant/blob/main/LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](https://github.com/hasura/ndc-qdrant/blob/main/README.md)

The Hasura Qdrant Connector allows for connecting to a Qdrant database to give you an instant GraphQL API on top of your
Qdrant data.

## Features

Below, you'll find a matrix of all supported features for the Qdrant connector:

| Feature                         | Supported | Notes                                                              |
| ------------------------------- | --------- | ------------------------------------------------------------------ |
| Native Queries + Logical Models | ❌        |                                                                    |
| Simple Object Query             | ✅        |                                                                    |
| Filter / Search                 | ✅        |                                                                    |
| Simple Aggregation              | ❌        |                                                                    |
| Sort                            | ❌        |                                                                    |
| Paginate                        | ✅        | Pagination offset field only works for documents with Integer ID's |
| Nested Objects                  | ✅        |                                                                    |
| Nested Arrays                   | ✅        |                                                                    |
| Nested Filtering                | ❌        |                                                                    |
| Nested Sorting                  | ❌        |                                                                    |
| Nested Relationships            | ❌        |                                                                    |
| Vector Search                   | ✅        |                                                                    |

## Build on Hasura DDN

[Get started](https://hasura.io/docs/3.0/how-to-build-with-ddn/with-qdrant) by connecting your Qdrant cluster to a Hasura DDN project.

## Fork the connector

You can fork the [connector's repo](https://github.com/hasura/ndc-qdrant) and iterate on it yourself.

## License

The Qdrant connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
