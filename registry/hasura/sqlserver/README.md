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
(Note: here and following we are naming the subgraph "my_subgraph" and the connector "ms_sql")

   ```bash
   ddn connector init ms_sql --subgraph my_subgraph/subgraph.yaml --hub-connector hasura/sqlserver --configure-port 8081 --add-to-compose-file compose.yaml
   ```

### 2. Add your SQLServer credentials

Add your credentials to `my_subgraph/connector/ms_sql/.env.local`

```env title="my_subgraph/connector/ms_sql/.env.local"
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://local.hasura.dev:4317
OTEL_SERVICE_NAME=my_subgraph_ms_sql
CONNECTION_URI=<YOUR_SQLSERVER_URL>
```

### 3. Introspect your Database

From the root of your project run:

```bash title="From the root of your project run:"
ddn connector introspect --connector my_subgraph/connector/ms_sql/connector.local.yaml
```

If you look at the `configuration.json` for your connector, you'll see metadata describing your SQL Server mappings.

### 4. Restart the services

Let's restart the docker compose services. Run the folowing from the root of your project:

```bash title="From the root of your project run:"
HASURA_DDN_PAT=$(ddn auth print-pat) docker compose up --build --watch
```

### 5. Create the Hasura metadata

In a new terminal tab from your project's root directory run:

```bash title="Run the following from the root of your project:"
ddn connector-link add ms_sql --subgraph my_subgraph/subgraph.yaml --configure-host http://local.hasura.dev:8081 --target-env-file my_subgraph/.env.my_subgraph.local
```

The above step will add the following env vars to the `.env.my_subgraph.local` file.

```env title="my_subgraph/.env.my_subgraph.local"
MY_SUBGRAPH_MS_SQL_READ_URL=http://local.hasura.dev:8081
MY_SUBGRAPH_MS_SQL_WRITE_URL=http://local.hasura.dev:8081
```

The generated file has two environment variables — one for reads and one for writes.
Each key is prefixed by the subgraph name, an underscore, and the name of the
connector.

### 6. Update the new DataConnectorLink object

Finally, now that our `DataConnectorLink` has the correct environment variables configured for the SQL Server connector,
we can run the update command to have the CLI look at the configuration JSON and transform it to reflect our database's
schema in `hml` format. From your project's root directory, run:

```bash title="From the root of your project, run:"
ddn connector-link update ms_sql --subgraph my_subgraph/subgraph.yaml --env-file my_subgraph/.env.my_subgraph.local
```

After this command runs, you can open your `my_subgraph/metadata/ms_sql.hml` file and see your metadata completely
scaffolded out for you 🎉

The schema of the database can be viewed at http://localhost:8081/schema.

### 7. Import _all_ your indices

You can do this with just one command. From your project's root directory, run:

```bash title="From the root of your project, run:"
ddn connector-link update ms_sql --subgraph my_subgraph/subgraph.yaml --env-file my_subgraph/.env.my_subgraph.local --add-all-resources
```

### 8. Create a supergraph build

Pass the `local` subcommand along with specifying the output directory as `./engine` in the root of the project. This
directory is used by the docker-compose file to serve the engine locally. From your project's root directory, run:

```bash title="From the root of your project, run:"
ddn supergraph build local --output-dir ./engine --subgraph-env-file my_subgraph:my_subgraph/.env.my_subgraph.local
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
