## Overview

The ClickHouse Native Data Connector allows for connecting to a ClickHouse instance giving you an instant GraphQL API on top of your ClickHouse data.
This uses the [Rust Data Connector SDK](https://github.com/hasura/ndc-hub#rusk-sdk) from the [Data connector Hub](https://github.com/hasura/ndc-hub) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

ClickHouse is a powerful open-source columnar database that offers a range of features designed for speed and efficiency in processing large volumes of data. ClickHouse is an excellent choice for a database when you are dealing with large volumes of data and require high-speed data retrieval, aggregation, and analysis. It's particularly well-suited for real-time analytics and handling time-series data, log data, or any scenario where read operations vastly outnumber writes. ClickHouse thrives in environments where query performance and the ability to generate reports quickly are critical, such as in financial analysis, IoT data management, and online analytical processing (OLAP). Furthermore, its column-oriented architecture makes it ideal for queries that need to scan large datasets but only access a subset of columns.

This native data connector implements the following Hasura Data Domain Specification features:

| Feature                                                                                                                             | Supported |
| ----------------------------------------------------------------------------------------------------------------------------------- | --------- |
| [Simple Queries](https://hasura.io/docs/3.0/graphql-api/queries/simple-queries/)                                                    | ✅        |
| [Nested Queries](https://hasura.io/docs/3.0/graphql-api/queries/nested-queries/)                                                    | ✅        |
| [Query Result Sorting](https://hasura.io/docs/3.0/graphql-api/queries/sorting/)                                                     | ✅        |
| [Query Result Pagination](https://hasura.io/docs/3.0/graphql-api/queries/pagination/)                                               | ✅        |
| [Multiple Query Arguments](https://hasura.io/docs/3.0/graphql-api/queries/multiple-arguments/)                                      | ✅        |
| [Multiple Queries in a Request](https://hasura.io/docs/3.0/graphql-api/queries/multiple-queries/)                                   | ✅        |
| [Variables, Aliases, Fragments, Directives](https://hasura.io/docs/3.0/graphql-api/queries/variables-aliases-fragments-directives/) | ✅        |
| [Query Filter: Value Comparison](https://hasura.io/docs/3.0/graphql-api/queries/filters/comparison-operators/)                      | ✅        |
| [Query Filter: Boolean Expressions](https://hasura.io/docs/3.0/graphql-api/queries/filters/boolean-operators/)                      | ✅        |
| [Query Filter: Text](https://hasura.io/docs/3.0/graphql-api/queries/filters/text-search-operators/)                                 | ✅        |
| [Query Filter: Nested Objects](https://hasura.io/docs/3.0/graphql-api/queries/filters/nested-objects/)                              | ✅        |

## Build on Hasura DDN

[Get started](https://hasura.io/docs/3.0/how-to-build-with-ddn/with-clickhouse) by connecting your ClickHouse instance to a Hasura DDN project.

## Fork the connector

You can fork the [connector's repo](https://github.com/hasura/ndc-clickhouse) and iterate on it yourself.
