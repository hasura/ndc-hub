# SQL Server Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/getting-started/overview/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-sqlserver-blue.svg?style=flat)](https://hasura.io/connectors/sqlserver)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](./readme.md)

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your data models in
Microsoft SQL Server. This connector supports SQL Server's functionalities listed in the table below, allowing for
efficient and scalable data operations. Additionally, users benefit from all the powerful features of Hasura’s Data
Delivery Network (DDN) platform, including query pushdown capabilities that delegate query operations to the database,
thereby enhancing query optimization and performance.

This connector is built using the [Rust Data Connector SDK](https://github.com/hasura/ndc-hub#rusk-sdk) and implements
the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [Connector information in the Hasura Hub](https://hasura.io/connectors/sqlserver)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0)

## Features

Below, you'll find a matrix of all supported features for the SQL Server connector:

| Feature                         | Supported | Notes                                |
|---------------------------------|-----------|--------------------------------------|
| Native Queries + Logical Models | ✅        |                                      |
| Native Mutations                | ✅        |                                      |
| Simple Object Query             | ✅        |                                      |
| Filter / Search                 | ✅        |                                      |
| Simple Aggregation              | ✅        |                                      |
| Sort                            | ✅        |                                      |
| Paginate                        | ✅        |                                      |
| Table Relationships             | ✅        |                                      |
| Views                           | ✅        |                                      |
| Remote Relationships            | ✅        |                                      |
| Custom Fields                   | ❌        |                                      |
| Mutations                       | ✅        | Only native mutations are suppported |
| Distinct                        | ✅        |                                      |
| Enums                           | ❌        |                                      |
| Naming Conventions              | ❌        |                                      |
| Default Values                  | ❌        |                                      |
| User-defined Functions          | ❌        |                                      |

## Before you get Started

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Install the [CLI](https://hasura.io/docs/3.0/cli/installation/)
3. Install the [Hasura VS Code extension](https://marketplace.visualstudio.com/items?itemName=HasuraHQ.hasura)
4. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
5. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

## Using the connector

To use the SQL Server connector, follow these steps in a Hasura project:
(Note: for more information on the following steps, please refer to the Postgres connector documentation [here](https://hasura.io/docs/3.0/getting-started/connect-to-data/connect-a-source))

### 1. Init the connector
(Note: here and following we are naming the subgraph "my_subgraph" and the connector "my_sql")

   ```bash
   ddn connector init my_sql --subgraph my_subgraph --hub-connector hasura/sqlserver
   ```

### 2. Add your SQLServer credentials:

Add your credentials to `my_subgraph/connector/my_sql/.env.local`

```env title="my_subgraph/connector/my_sql/.env.local"
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://local.hasura.dev:4317
OTEL_SERVICE_NAME=my_subgraph_my_sql
CONNECTION_URI=<YOUR_SQLSERVER_URL>
```

### 3. Intropsect your indices

```bash title="From the root of your project run:"
ddn connector introspect --connector my_subgraph/connector/my_sql/connector.yaml
```

If you look at the `configuration.json` for your connector, you'll see metadata describing your SQL Server mappings.

### 4. Create the Hasura metadata

```bash title="Run the following from the root of your project:"
ddn connector-link add my_sql --subgraph my_subgraph
```

The generated file has two environment variables — one for reads and one for writes — that you'll need to add to your
subgraph's `.env.my_subgraph` file. Each key is prefixed by the subgraph name, an underscore, and the name of the
connector. Ensure the port value matches what is published in your connector's docker compose file.

```env title="my_subgraph/.env.my_subgraph"
MY_SUBGRAPH_MY_SQL_READ_URL=http://local.hasura.dev:8081
MY_SUBGRAPH_MY_SQL_WRITE_URL=http://local.hasura.dev:8081
```

### 5. Start the connector's docker compose

Let's start our connector's docker compose file.

```bash title="Run the following from the connector's subdirectory inside a subgraph:"
docker compose -f docker-compose.my_sql.yaml up
```

This starts our SQL Server connector on the specified port. We can navigate to the following address, with the port
modified, to see the schema of our SQL Server data source:

```bash
http://localhost:8081/schema
```

### 6. Include the connector in your docker compose

Kill the connector by pressing `CTRL+C` in the terminal tab in which the connector is running.

Then, add the following inclusion to the docker compose `docker-compose.hasura.yaml` in your project's root directory, taking care to modify the
subgraph's name.

```yaml title="docker-compose.hasura.yaml"
include:
  - path: my_subgraph/connector/my_sql/docker-compose.my_sql.yaml
```

Now, whenever running the following, you'll bring up the GraphQL engine, observability tools, and any connectors you've
included:

```bash title="From the root of your project, run:"
HASURA_DDN_PAT=$(ddn auth print-pat) docker compose -f docker-compose.hasura.yaml watch
```

### 7. Update the new DataConnectorLink object

Finally, now that our `DataConnectorLink` has the correct environment variables configured for the SQL Server connector,
we can run the update command to have the CLI look at the configuration JSON and transform it to reflect our database's
schema in `hml` format. In a new terminal tab, run:

```bash title="From the root of your project, run:"
ddn connector-link update my_sql --subgraph my_subgraph
```

After this command runs, you can open your `my_subgraph/metadata/my_sql.hml` file and see your metadata completely
scaffolded out for you 🎉

### 8. Import _all_ your indices

You can do this in one convenience command.

```bash title="From the root of your project, run:"
ddn connector-link update my_sql --subgraph my_subgraph --add-all-resources
```

### 9. Create a supergraph build

Pass the `local` subcommand along with specifying the output directory as `./engine` in the root of the project. This
directory is used by the docker-compose file to serve the engine locally:

```bash title="From the root of your project, run:"
ddn supergraph build local --output-dir ./engine
```

You can now navigate to
[`https://console.hasura.io/local/graphql?url=http://localhost:3000`](https://console.hasura.io/local/graphql?url=http://localhost:3000)
and interact with your API using the Hasura Console.

## Documentation

View the full documentation for the ndc-sqlserver connector [here](https://github.com/hasura/ndc-sqlserver/blob/main/docs/readme.md).

## Contributing

We're happy to receive any contributions from the community. Please refer to our [development guide](https://github.com/hasura/ndc-sqlserver/blob/main/docs/development.md).

## License

The Hasura SQL Server connector is available under the [Apache License
2.0](https://www.apache.org/licenses/LICENSE-2.0).
