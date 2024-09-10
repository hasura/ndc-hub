## Overview

The Calcite Native Data Connector allows for connecting to a Calcite instance giving you an instant GraphQL API on top of your Calcite data.
This uses the [Rust Data Connector SDK](https://github.com/hasura/ndc-hub#rusk-sdk) from the [Data connector Hub](https://github.com/hasura/ndc-hub) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

Calcite is a powerful open-source columnar database that offers a range of features designed for speed and efficiency in processing large volumes of data. Calcite is an excellent choice for a database when you are dealing with large volumes of data and require high-speed data retrieval, aggregation, and analysis. It's particularly well-suited for real-time analytics and handling time-series data, log data, or any scenario where read operations vastly outnumber writes. Calcite thrives in environments where query performance and the ability to generate reports quickly are critical, such as in financial analysis, IoT data management, and online analytical processing (OLAP). Furthermore, its column-oriented architecture makes it ideal for queries that need to scan large datasets but only access a subset of columns. 

## Connect to Hasura

Please refer to the [Getting Started - Create an API](https://hasura.io/docs/3.0/getting-started/create-a-project) documentation if you get stuck during any of the steps outlined below.

### Prerequisites
1. Install the [new Hasura CLI](https://hasura.io/docs/3.0/cli/installation) — to quickly and easily create and manage your Hasura projects and builds.
2. (recommended) Install the [Hasura VS Code extension](https://marketplace.visualstudio.com/items?itemName=HasuraHQ.hasura) — with support for other editors coming soon!
3. Create a [Calcite account](https://Calcite.cloud/signUp?loc=nav-get-started) if you don't already have one.
4. Make sure to make your Calcite service open to the public or add Hasura's IP to the allowlist.

### Create Project and Connect Calcite

Login to Hasura Cloud with the CLI

```
ddn login
```

Create a new project using the [create project](https://hasura.io/docs/3.0/cli/commands/create-project/) command in the CLI and change to the new directory that was generated.

```
ddn create project --dir ./my-first-supergraph
cd my-first-supergraph
```

Run the add [connector-manifest](https://hasura.io/docs/3.0/cli/commands/add-connector-manifest/) command to create a connector for Calcite in your project. 

```
ddn add connector-manifest Calcite_connector --subgraph app --hub-connector hasura/Calcite --type cloud
```

Add values for your Calcite username, password, and connection string to corresponding definition found in: `app/Calcite/connector/Calcite_connector.build.hml`

```
kind: ConnectorManifest
version: v1
spec:
  supergraphManifests:
    - base
definition:
  name: Calcite_connector
  type: cloud
  connector:
    type: hub
    name: hasura/Calcite:v0.2.5
  deployments:
    - context: .
      env:
        CALCITE_PASSWORD:
          value: ""
        CALCITE_URL:
          value: ""
        CALCITE_USERNAME:
          value: ""
```

Note: You can also use environment variables for these values. Please refer to our [Getting Started - Add a connector manifest](https://hasura.io/docs/3.0/cli/commands/add-connector-manifest/) for more details.

### Update Connector, Track Models and Build

At this point you can either run the [dev](https://hasura.io/docs/3.0/cli/commands/dev/) mode to watch your project and create new builds as changes are made to your metadata using [Hasura’s LSP](https://hasura.io/docs/3.0/glossary/#lsp-language-server-protocol) and [VSCode extension](https://marketplace.visualstudio.com/items?itemName=HasuraHQ.hasura).

```
ddn dev
```

Alternatively, you can run the following commands to add specific models, in this example the `Trips` table and `MonthlyRevenue` view if added the view following the steps mentioned above.
```
ddn update connector-manifest Calcite_connector
ddn update data-connector-link Calcite_connector
ddn add model --data-connector-link Calcite_connector --name Trips
ddn add model --data-connector-link Calcite_connector --name MonthlyRevenue
ddn build supergraph-manifest
```

You are now ready to start using your API! During the previous step the console will return some information including the Console URL. Load this link in your browser to explore the API you have created for your Calcite database. The UI will resemble something like this.
