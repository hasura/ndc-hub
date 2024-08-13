# Hasura Qdrant Connector
<a href="https://qdrant.tech/"><img src="https://github.com/hasura/ndc-qdrant/blob/main/docs/logo.png" align="right" width="200"></a>

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/connectors/qdrant)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-qdrant-blue.svg?style=flat)](https://hasura.io/connectors/qdrant)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://github.com/hasura/ndc-qdrant/blob/main/LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](https://github.com/hasura/ndc-qdrant/blob/main/README.md)

The Hasura Qdrant Connector allows for connecting to a Qdrant database to give you an instant GraphQL API on top of your Qdrant data.

This connector is built using the [Typescript Data Connector SDK](https://github.com/hasura/ndc-sdk-typescript) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

* [Connector information in the Hasura Hub](https://hasura.io/connectors/qdrant)
* [Hasura V3 Documentation](https://hasura.io/docs/3.0/index/)

## Features

Below, you'll find a matrix of all supported features for the Qdrant connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models | ❌     |       |
| Simple Object Query             | ✅     |       |
| Filter / Search                 | ✅     |       |
| Simple Aggregation              | ❌     |       |
| Sort                            | ❌     |       |
| Paginate                        | ✅     | Pagination offset field only works for documents with Integer ID's       |
| Nested Objects                  | ✅     |       |
| Nested Arrays                   | ✅     |       |
| Nested Filtering                | ❌     |       |
| Nested Sorting                  | ❌     |       |
| Nested Relationships            | ❌     |       |
| Vector Search                   | ✅     |       |

## Before you get Started

[Prerequisites or recommended steps before using the connector.]

1. The [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and [Docker](https://docs.docker.com/engine/install/) installed
2. A [supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
3. A [subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)
4. Have a [Qdrant](https://qdrant.tech/) hosted database, or a locally running Qdrant database — for supplying data to your API.

The steps below explain how to Initialize and configure a connector for local development. You can learn how to deploy a
connector — after it's been configured — [here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the Qdrant connector

### Step 1: Authenticate your CLI session

```bash
ddn auth login
```

### Step 2: Initialize the connector

```bash
ddn connector init qdrant  --subgraph my_subgraph  --hub-connector hasura/qdrant
```

In the snippet above, we've used the subgraph `my_subgraph` as an example; however, you should change this
value to match any subgraph which you've created in your project.

### Step 3: Modify the connector's port

When you initialized your connector, the CLI generated a set of configuration files, including a Docker Compose file for
the connector. Typically, connectors default to port `8080`. Each time you add a connector, we recommend incrementing the published port by one to avoid port collisions.

As an example, if your connector's configuration is in `my_subgraph/connector/qdrant/docker-compose.qdrant.yaml`, you can modify the published port to reflect a value that isn't currently being used by any other connectors:

```yaml
ports:
  - mode: ingress
    target: 8080
    published: "8082"
    protocol: tcp
```

### Step 4: Add environment variables

Now that our connector has been scaffolded out for us, we need to provide a connection string so that the data source can be introspected and the boilerplate configuration can be taken care of by the CLI.

The CLI has provided an `.env.local` file for our connector in the `my_subgraph/connector/qdrant` directory. We can add a key-value pair of `QDRANT_URL` along with the connection string itself to this file, and our connector will use this to connect to our database. If the Qdrant database has an API key you can also provide the environment variable for the `QDRANT_API_KEY` which allows the connector to authenticate.


The file, after adding the `QDRANT_URL`, should look like this example if connecting to a Qdrant hosted database instance:

```env
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://local.hasura.dev:4317
OTEL_SERVICE_NAME=my_subgraph_qdrant
QDRANT_URL=https://7312d6c4-3f6c-432c-987c-34d7d96428ef.us-east4-0.gcp.cloud.qdrant.io
QDRANT_API_KEY=Ch8I...
```

To connect to a locally running Qdrant instance you can then point the `QDRANT_URL` to the local database. Assuming the Qdrant database is running on port 6333 without any API key, you should be able to use this example:

```env
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://local.hasura.dev:4317
OTEL_SERVICE_NAME=my_subgraph_qdrant
QDRANT_URL=http://local.hasura.dev:6333
```

### Step 5: Introspect your data source

With the connector configured, we can now use the CLI to introspect our database and create a source-specific configuration file for our connector.

```bash
ddn connector introspect --connector my_subgraph/connector/qdrant/connector.local.yaml
```

### Step 6. Create the Hasura metadata

Hasura DDN uses a concept called "connector linking" to take [NDC-compliant](https://github.com/hasura/ndc-spec)
configuration JSON files for a data connector and transform them into an `hml` (Hasura Metadata Language) file as a
[`DataConnectorLink` metadata object](https://hasura.io/docs/3.0/supergraph-modeling/data-connectors#dataconnectorlink-dataconnectorlink).

Basically, metadata objects in `hml` files define our API.

First we need to create this `hml` file with the `connector-link add` command and then convert our configuration files
into `hml` syntax and add it to this file with the `connector-link update` command.

Let's name the `hml` file the same as our connector, `qdrant`:

```bash
ddn connector-link add qdrant --subgraph my_subgraph
```

The new file is scaffolded out at `my_subgraph/metadata/qdrant/qdrant.hml`.

### Step 7. Update the environment variables

The generated file has two environment variables — one for reads and one for writes — that you'll need to add to your subgraph's `.env.my_subgraph` file. Each key is prefixed by the subgraph name, an underscore, and the name of the connector. Ensure the port value matches what is published in your connector's docker compose file.

As an example:

```env
MY_SUBGRAPH_QDRANT_READ_URL=http://local.hasura.dev:<port>
MY_SUBGRAPH_QDRANT_WRITE_URL=http://local.hasura.dev:<port>
```

These values are for the connector itself and utilize `local.hasura.dev` to ensure proper resolution within the docker container.

### Step 8. Start the connector's Docker Compose

Let's start our connector's Docker Compose file by running the following from inside the connector's subgraph:

```bash
docker compose -f docker-compose.qdrant.yaml up
```

### Step 9. Update the new `DataConnectorLink` object

Finally, now that our `DataConnectorLink` has the correct environment variables configured for the connector, we can run the update command to have the CLI look at the configuration JSON and transform it to reflect our database's schema in `hml` format. In a new terminal tab, run:

```bash
ddn connector-link update qdrant --subgraph my_subgraph
```

After this command runs, you can open your `my_subgraph/metadata/qdrant.hml` file and see your metadata completely
scaffolded out for you 🎉

## Documentation

View the full documentation for the Qdrant connector [here](https://github.com/hasura/ndc-qdrant/blob/main/docs/index.md).

## Contributing

Check out our [contributing guide](https://github.com/hasura/ndc-qdrant/blob/main/docs/contributing.md) for more details.

## License

The Qdrant connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
