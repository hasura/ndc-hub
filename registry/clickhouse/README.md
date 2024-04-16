## Overview

The Clickhouse Native Data Connector allows for connecting to a Clickhouse instance giving you an instant GraphQL API on top of your Clickhouse data.
This uses the [Rust Data Connector SDK](https://github.com/hasura/ndc-hub#rusk-sdk) from the [Data connector Hub](https://github.com/hasura/ndc-hub) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

ClickHouse is a powerful open-source columnar database that offers a range of features designed for speed and efficiency in processing large volumes of data. ClickHouse is an excellent choice for a database when you are dealing with large volumes of data and require high-speed data retrieval, aggregation, and analysis. It's particularly well-suited for real-time analytics and handling time-series data, log data, or any scenario where read operations vastly outnumber writes. ClickHouse thrives in environments where query performance and the ability to generate reports quickly are critical, such as in financial analysis, IoT data management, and online analytical processing (OLAP). Furthermore, its column-oriented architecture makes it ideal for queries that need to scan large datasets but only access a subset of columns. 

## Connect to Hasura

Please refer to the [Getting Started - Create an API](https://hasura.io/docs/3.0/getting-started/create-a-project) documentation if you get stuck during any of the steps outlined below.

### Prerequisites
1. Install the [new Hasura CLI](https://hasura.io/docs/3.0/cli/installation) — to quickly and easily create and manage your Hasura projects and builds.
2. (recommended) Install the [Hasura VS Code extension](https://marketplace.visualstudio.com/items?itemName=HasuraHQ.hasura) — with support for other editors coming soon!
3. Create a [Clickhouse account](https://clickhouse.cloud/signUp?loc=nav-get-started) if you don't already have one.
4. Make sure to make your Clickhouse service open to the public or add Hasura's IP to the allowlist.

### Create Project and Connect Clickhouse

Login to Hasura Cloud with the CLI

```
ddn login
```

Create a new project using the [create project](https://hasura.io/docs/3.0/cli/commands/create-project/) command in the CLI and change to the new directory that was generated.

```
ddn create project --dir ./my-first-supergraph
cd my-first-supergraph
```

Run the add [connector-manifest](https://hasura.io/docs/3.0/cli/commands/add-connector-manifest/) command to create a connector for Clickhouse in your project. 

```
ddn add connector-manifest clickhouse_connector --subgraph app --hub-connector hasura/clickhouse --type cloud
```

Add values for your Clickhouse username, password, and connection string to corresponding definition found in: `app/clickhouse/connector/clickhouse_connector.build.hml`

```
kind: ConnectorManifest
version: v1
spec:
  supergraphManifests:
    - base
definition:
  name: clickhouse_connector
  type: cloud
  connector:
    type: hub
    name: hasura/clickhouse:v0.2.5
  deployments:
    - context: .
      env:
        CLICKHOUSE_PASSWORD:
          value: ""
        CLICKHOUSE_URL:
          value: ""
        CLICKHOUSE_USERNAME:
          value: ""
```

Note: You can also use environment variables for these values. Please refer to our [Getting Started - Add a connector manifest](https://hasura.io/docs/3.0/cli/commands/add-connector-manifest/) for more details.

### Update Connector, Track Models and Build

At this point you can either run the <code>[dev](https://hasura.io/docs/3.0/cli/commands/dev/)</code> mode to watch your project and create new builds as changes are made to your metadata using [Hasura’s LSP](https://hasura.io/docs/3.0/glossary/#lsp-language-server-protocol) and [VSCode extension](https://marketplace.visualstudio.com/items?itemName=HasuraHQ.hasura).

```
ddn dev
```

Alternatively, you can run the following commands to add specific models, in this example the `Trips` table and `MonthlyRevenue` view if added the view following the steps mentioned above.
```
ddn update connector-manifest clickhouse_connector
ddn update data-connector-link clickhouse_connector
ddn add model --data-connector-link clickhouse_connector --name Trips
ddn add model --data-connector-link clickhouse_connector --name MonthlyRevenue
ddn build supergraph-manifest
```

You are now ready to start using your API! During the previous step the console will return some information including the Console URL. Load this link in your browser to explore the API you have created for your Clickhouse database. The UI will resemble something like this.