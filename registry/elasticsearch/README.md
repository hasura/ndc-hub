# Elasticsearch Connector

<a href="https://hasura.io/"><img src="./logo.png" align="right" width="200"></a>

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/latest/connectors/elasticsearch/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-elasticsearch-blue.svg?style=flat)](https://hasura.io/connectors/ndc-elasticsearch)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](./readme.md)

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your documents in Elasticsearch. This connector supports Elasticsearch functionalities listed in the table below, allowing for efficient and scalable data operations. Additionally, you will benefit from all the powerful features of Hasura’s Data Delivery Network (DDN) platform, including query pushdown capabilities that delegate all query operations to the Elasticsearch, thereby enhancing query optimization and performance.

This connector is built using the [Go Data Connector SDK](https://github.com/hasura/ndc-sdk-go) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [Connector information in the Hasura Hub](https://hasura.io/connectors/elasticsearch)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0)

## Features

Below, you'll find a matrix of all supported features for the Elasticsearch connector:

<!-- DocDB matrix -->

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models | ❌         |       |
| Simple Object Query             | ✅         |       |
| Filter / Search                 | ✅         |       |
| Simple Aggregation              | ✅         |       |
| Sort                            | ✅         |       |
| Paginate                        | ✅         |       |
| Nested Objects                  | ✅         |       |
| Nested Arrays                   | ✅         |       |
| Nested Filtering                | ❌         |       |
| Nested Sorting                  | ❌         |       |
| Nested Relationships            | ❌         |       |


## Before you get Started

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Install the [CLI](https://hasura.io/docs/3.0/cli/installation/)
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

## Using the connector

To use the Elasticsearch connector, follow these steps in a Hasura project:
(Note: for more information on the following steps, please refer to the Postgres connector documentation [here](https://hasura.io/docs/3.0/getting-started/connect-to-data/connect-a-source))


### 1. Init the connector
(Note: here and following we are naming the subgraph "my_subgraph" and the connector "my_elastic")

   ```bash
   ddn connector init my_elastic --subgraph my_subgraph --hub-connector hasura/elasticsearch
   ```

### 2. Add your Elasticsearch credentials:

```env title="my_subgraph/connector/my_elastic/.env.local"
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://local.hasura.dev:4317
OTEL_SERVICE_NAME=my_subgraph_my_elastic
ELASTICSEARCH_URL=<YOUR_ELASTICSEARCH_URL>
ELASTICSEARCH_USERNAME=<YOUR_ELASTICSEARCH_USERNAME>
ELASTICSEARCH_PASSWORD=<YOUR_ELASTICSEARCH_PASSWORD>
```

To configure the connector, the following environment variables need to be set:

| Environment Variable          | Description                                                                                                     | Required | Example Value                                                  |
| ----------------------------- | --------------------------------------------------------------------------------------------------------------- | -------- | -------------------------------------------------------------- |
| `ELASTICSEARCH_URL`           | The comma-separated list of Elasticsearch host addresses for connection                                         | Yes      | `https://example.es.gcp.cloud.es.io:9200`                      |
| `ELASTICSEARCH_USERNAME`      | The username for authenticating to the Elasticsearch cluster                                                    | Yes      | `admin`                                                        |
| `ELASTICSEARCH_PASSWORD`      | The password for the Elasticsearch user account                                                                 | Yes      | `default`                                                      |
| `ELASTICSEARCH_API_KEY`       | The Elasticsearch API key for authenticating to the Elasticsearch cluster                                       | No       | `ABCzYWk0NEI0aDRxxxxxxxxxx1k6LWVQa2gxMUpRTUstbjNwTFIzbGoyUQ==` |
| `ELASTICSEARCH_CA_CERT_PATH`  | The path to the Certificate Authority (CA) certificate for verifying the Elasticsearch server's SSL certificate | No       | `/etc/connector/cacert.pem`                                    |
| `ELASTICSEARCH_INDEX_PATTERN` | The pattern for matching Elasticsearch indices, potentially including wildcards, used by the connector          | No       | `hasura*`                                                      |


### 3. Intropsect your indices

```bash title="From the root of your project run:"
ddn connector introspect --connector my_subgraph/connector/my_elastic/connector.yaml
```

If you look at the `configuration.json` for your connector, you'll see metadata describing your Elasticsearch mappings.

### 4. Create the Hasura metadata

```bash title="Run the following from the root of your project:"
ddn connector-link add my_elastic --subgraph my_subgraph
```

The generated file has two environment variables — one for reads and one for writes — that you'll need to add to your
subgraph's `.env.my_subgraph` file. Each key is prefixed by the subgraph name, an underscore, and the name of the
connector. Ensure the port value matches what is published in your connector's docker compose file.

```env title="my_subgraph/.env.my_subgraph"
MY_SUBGRAPH_MY_ELASTIC_READ_URL=http://local.hasura.dev:8081
MY_SUBGRAPH_MY_ELASTIC_WRITE_URL=http://local.hasura.dev:8081
```

### 5. Start the connector's docker compose

Let's start our connector's docker compose file.

```bash title="Run the following from the connector's subdirectory inside a subgraph:"
docker compose -f docker-compose.my_elastic.yaml up
```

This starts our PostgreSQL connector on the specified port. We can navigate to the following address, with the port
modified, to see the schema of our Elasticsearch data source:

```bash
http://localhost:8081/schema
```

### 6. Include the connector in your docker compose

Kill the connector by pressing `CTRL+C` in the terminal tab in which the connector is running.

Then, add the following inclusion to the docker compose in your project's root directory, taking care to modify the
subgraph's name.

```yaml title="docker-compose.hasura.yaml"
include:
  - path: my_subgraph/connector/my_elastic/docker-compose.my_elastic.yaml
```

Now, whenever running the following, you'll bring up the GraphQL engine, observability tools, and any connectors you've
included:

```bash title="From the root of your project, run:"
HASURA_DDN_PAT=$(ddn auth print-pat) docker compose -f docker-compose.hasura.yaml watch
```

### 7. Update the new DataConnectorLink object

Finally, now that our `DataConnectorLink` has the correct environment variables configured for the Elasticsearch connector,
we can run the update command to have the CLI look at the configuration JSON and transform it to reflect our database's
schema in `hml` format. In a new terminal tab, run:

```bash title="From the root of your project, run:"
ddn connector-link update my_elastic --subgraph my_subgraph
```

After this command runs, you can open your `my_subgraph/metadata/my_elastic.hml` file and see your metadata completely
scaffolded out for you 🎉

### 8. Import _all_ your indices

You can do this in one convenience command.

```bash title="From the root of your project, run:"
ddn connector-link update my_elastic --subgraph my_subgraph --add-all-resources
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


## Contributing

Check out our [contributing guide](./docs/contributing.md) for more details.

## License

The Elasticsearch connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
