# Azure Cosmos DB for NoSQL Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/latest/connectors/azure-cosmos/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-azure--cosmos-blue.svg?style=flat)](https://hasura.io/connectors/azure-cosmos)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](./readme.md)

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your data models in Azure Cosmos DB for NoSQL Database containers. This connector supports Azure Cosmos DB for NoSQL's functionalities listed in the table below, allowing for efficient and scalable data operations.

This connector is built using the [TypeScript Data Connector SDK](https://github.com/hasura/ndc-sdk-typescript) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [Connector information in the Hasura Hub](https://hasura.io/connectors/azure-cosmos)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0)

## Features

Below, you'll find a matrix of all supported features for the Azure Cosmos DB for NoSQL connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models |    ✅     |       |
| Simple Object Query             |    ✅     |       |
| Filter / Search                 |    ✅     |       |
| Simple Aggregation              |    ✅     |       |
| Sort                            |    ✅     |       |
| Paginate                        |    ✅     |       |
| Nested Objects                  |    ✅     |       |
| Nested Arrays                   |    ✅     |       |
| Nested Filtering                |    ❌     |       |
| Nested Sorting                  |    ❌     |       |
| Nested Relationships            |    ❌     |       |


## Before you get Started

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Install the [CLI](https://hasura.io/docs/3.0/cli/installation/)
3. Install the [Hasura VS Code extension](https://marketplace.visualstudio.com/items?itemName=HasuraHQ.hasura)
4. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
5. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

## Using the connector

To use the Azure Cosmos DB for NoSQL connector, follow these steps in a Hasura project:
(Note: for more information on the following steps, please refer to the Postgres connector documentation [here](https://hasura.io/docs/3.0/getting-started/connect-to-data/connect-a-source))


### 1. Init the connector
(Note: here and following we are naming the subgraph "my_subgraph" and the connector "my_azure_cosmos")

   ```bash
   ddn connector init my_azure_cosmos --subgraph my_subgraph --hub-connector hasura/azure-cosmos
   ```

### 2. Add your Azure Cosmos DB for NoSQL credentials

Add you credentials to `my_subgraph/connector/my_azure_cosmos/.env.local`

```env title="my_subgraph/connector/my_azure_cosmos/.env.local"
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://local.hasura.dev:4317
OTEL_SERVICE_NAME=my_subgraph_my_azure_cosmos
AZURE_COSMOS_DB_NAME= <YOUR_AZURE_DB_NAME>
AZURE_COSMOS_ENDPOINT= <YOUR_AZURE_COSMOS_ENDPOINT>
AZURE_COSMOS_KEY= <YOUR_AZURE_COSMOS_KEY>
AZURE_COSMOS_NO_OF_ROWS_TO_FETCH= <NO-OF-ROWS-TO-FETCH>
```

Note: `AZURE_COSMOS_CONNECTOR_NO_OF_ROWS_TO_FETCH` is an optional field, with 100 rows to be fetched by default.

### 3. Introspect your indices

From the root of your project run:

```bash title="From the root of your project run:"
ddn connector introspect --connector my_subgraph/connector/my_azure_cosmos/connector.yaml
```

If you look at the `config.json` for your connector, you'll see metadata describing your Azure Cosmos DB for NoSQL mappings.

### 4. Create the Hasura metadata

Run the following from the root of your project:

```bash title="Run the following from the root of your project:"
ddn connector-link add my_azure_cosmos --subgraph my_subgraph
```

The generated file has two environment variables — one for reads and one for writes — that you'll need to add to your
subgraph's `.env.my_subgraph` file. Each key is prefixed by the subgraph name, an underscore, and the name of the
connector. Ensure the port value matches what is published in your connector's docker compose file.

```env title="my_subgraph/.env.my_subgraph"
MY_SUBGRAPH_MY_AZURE_COSMOS_READ_URL=http://local.hasura.dev:8081
MY_SUBGRAPH_MY_AZURE_COSMOS_WRITE_URL=http://local.hasura.dev:8081
```

### 5. Start the connector's docker compose

Let's start our connector's docker compose file. Run the following from the connector's subdirectory inside a subgraph:

```bash title="Run the following from the connector's subdirectory inside a subgraph:"
docker compose -f docker-compose.my_azure_cosmos.yaml up
```

This starts our Azure Cosmos DB for NoSQL connector on the specified port. We can navigate to the following address, with the port
modified, to see the schema of our Azure Cosmos DB for NoSQL source:

```bash
http://localhost:8081/schema
```

### 6. Include the connector in your docker compose

Kill the connector by pressing `CTRL+C` in the terminal tab in which the connector is running.

Then, add the following inclusion to the docker compose `docker-compose.hasura.yaml` in your project's root directory, taking care to modify the
subgraph's name.

```yaml title="docker-compose.hasura.yaml"
include:
  - path: my_subgraph/connector/my_azure_cosmos/docker-compose.my_azure_cosmos.yaml
```

Now, whenever running the following, you'll bring up the GraphQL engine, observability tools, and any connectors you've
included. From the root of your project, run:

```bash title="From the root of your project, run:"
HASURA_DDN_PAT=$(ddn auth print-pat) docker compose -f docker-compose.hasura.yaml watch
```

### 7. Update the new DataConnectorLink object

Finally, now that our `DataConnectorLink` has the correct environment variables configured for the Azure Cosmos DB for NoSQL connector,
we can run the update command to have the CLI look at the configuration JSON and transform it to reflect our database's
schema in `hml` format. In a new terminal tab from the root of your project, run:

```bash title="From the root of your project, run:"
ddn connector-link update my_azure_cosmos --subgraph my_subgraph
```

After this command runs, you can open your `my_subgraph/metadata/my_azure_cosmos.hml` file and see your metadata completely
scaffolded out for you 🎉

### 8. Import _all_ your indices

You can do this in one convenience command. From the root of your project, run:

```bash title="From the root of your project, run:"
ddn connector-link update my_azure_cosmos --subgraph my_subgraph --add-all-resources
```

### 9. Create a supergraph build

Pass the `local` subcommand along with specifying the output directory as `./engine` in the root of the project. This
directory is used by the docker-compose file to serve the engine locally. From the root of your project, run:

```bash title="From the root of your project, run:"
ddn supergraph build local --output-dir ./engine
```

You can now navigate to
[`https://console.hasura.io/local/graphql?url=http://localhost:3000`](https://console.hasura.io/local/graphql?url=http://localhost:3000)
and interact with your API using the Hasura Console.

## Contributing

We're happy to receive any contributions from the community. Please refer to our [development guide](https://github.com/hasura/ndc-azure-cosmos-connector/blob/main/docs/development.md).

## License

The Hasura Azure Cosmos DB for NoSQL connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
