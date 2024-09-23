## Overview

The ClickHouse Native Data Connector allows for connecting to a ClickHouse instance giving you an instant GraphQL API on top of your ClickHouse data.
This uses the [Rust Data Connector SDK](https://github.com/hasura/ndc-hub#rusk-sdk) from the [Data connector Hub](https://github.com/hasura/ndc-hub) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

ClickHouse is a powerful open-source columnar database that offers a range of features designed for speed and efficiency in processing large volumes of data. ClickHouse is an excellent choice for a database when you are dealing with large volumes of data and require high-speed data retrieval, aggregation, and analysis. It's particularly well-suited for real-time analytics and handling time-series data, log data, or any scenario where read operations vastly outnumber writes. ClickHouse thrives in environments where query performance and the ability to generate reports quickly are critical, such as in financial analysis, IoT data management, and online analytical processing (OLAP). Furthermore, its column-oriented architecture makes it ideal for queries that need to scan large datasets but only access a subset of columns.

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)
5. Create a [ClickHouse account](https://clickhouse.cloud/signUp?loc=nav-get-started) if you don't already have one.
6. Make sure to make your ClickHouse service open to the public or add Hasura's IP to the allowlist.

## Using the ClickHouse connector

Check out the [Hasura docs here](https://hasura.io/docs/3.0/getting-started/build/connect-to-data/connect-a-source/?db=ClickHouse) to get started with the ClickHouse connector.
