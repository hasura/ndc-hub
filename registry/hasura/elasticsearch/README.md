# Elasticsearch Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-elasticsearch-blue.svg?style=flat)](https://hasura.io/connectors/elasticsearch)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://www.apache.org/licenses/LICENSE-2.0)

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your documents in
Elasticsearch. This connector supports Elasticsearch functionalities listed in the table below, allowing for efficient
and scalable data operations. Additionally, you will benefit from all the powerful features of Hasura’s Data Delivery
Network (DDN) platform, including query pushdown capabilities that delegate all query operations to the Elasticsearch,
thereby enhancing query optimization and performance.

## Features

Below, you'll find a matrix of all supported features for the Elasticsearch connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models | ❌        |       |
| Simple Object Query             | ✅        |       |
| Filter / Search                 | ✅        |       |
| Simple Aggregation              | ✅        |       |
| Sort                            | ✅        |       |
| Paginate                        | ✅        |       |
| Relationships                   | ✅        |       |
| Nested Objects                  | ✅        |       |
| Nested Arrays                   | ✅        |       |
| Nested Filtering                | ✅        |       |
| Nested Sorting                  | ❌        |       |
| Nested Relationships            | ❌        |       |

> [!Note]
>
> - **Relationships** are currently implemented via `top_hits` operator. That operator has a default maximum result size limit of 100 rows. This is what the connector operates on. If you give the connector a higher limit, it will change that to 100 for compliance with the database. Also, since the returned result will contain only 100 rows per bucket, it may not represent the whole result.

## Build on Hasura DDN

[Get started](https://hasura.io/docs/3.0/how-to-build-with-ddn/with-elasticsearch) by connecting your Elasticsearch instance to a Hasura DDN project.

## Fork the connector

You can fork the [connector's repo](https://github.com/hasura/ndc-elasticsearch) and iterate on it yourself.

## License

The Elasticsearch connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
