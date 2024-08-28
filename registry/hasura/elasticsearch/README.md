# Elasticsearch Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-elasticsearch-blue.svg?style=flat)](https://hasura.io/connectors/elasticsearch)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://www.apache.org/licenses/LICENSE-2.0)

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your documents in Elasticsearch. This connector supports Elasticsearch functionalities listed in the table below, allowing for efficient and scalable data operations. Additionally, you will benefit from all the powerful features of Hasura’s Data Delivery Network (DDN) platform, including query pushdown capabilities that delegate all query operations to the Elasticsearch, thereby enhancing query optimization and performance.

This connector is built using the [Go Data Connector SDK](https://github.com/hasura/ndc-sdk-go) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [See the listing in the Hasura Hub](https://hasura.io/connectors/elasticsearch)
- [Hasura DDN Documentation](https://hasura.io/docs/3.0)
- [Hasura DDN Quickstart](https://hasura.io/docs/3.0/getting-started/quickstart)
- [GraphQL on Elasticsearch](https://hasura.io/graphql/database/elasticsearch)

Docs for the Elasticsearch data connector:

- [Architecture](https://github.com/hasura/ndc-elasticsearch/blob/main/docs/architecture.md)
- [Code of Conduct](https://github.com/hasura/ndc-elasticsearch/blob/main/docs/code-of-conduct.md)
- [Contributing](https://github.com/hasura/ndc-elasticsearch/blob/main/docs/contributing.md)
- [Configuration](https://github.com/hasura/ndc-elasticsearch/blob/main/docs/configuration.md)
- [Development](https://github.com/hasura/ndc-elasticsearch/blob/main/docs/development.md)
- [Security](https://github.com/hasura/ndc-elasticsearch/blob/main/docs/security.md)
- [Support](https://github.com/hasura/ndc-elasticsearch/blob/main/docs/support.md)

## Features

Below, you'll find a matrix of all supported features for the Elasticsearch connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models | ✅        |       |
| Simple Object Query             | ✅        |       |
| Filter / Search                 | ✅        |       |
| Simple Aggregation              | ✅        |       |
| Sort                            | ✅        |       |
| Paginate                        | ✅        |       |
| Nested Objects                  | ✅        |       |
| Nested Arrays                   | ✅        |       |
| Nested Filtering                | ✅        |       |
| Nested Sorting                  | ❌        |       |
| Nested Relationships            | ❌        |       |

## Before you get Started

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

## Using the Elasticsearch connector

### Step 1: Authenticate your CLI session

```bash
ddn auth login
```

### Step 2: Configure the connector

Once you have an initialized supergraph and subgraph, run the initialization command in interactive mode while
providing a name for the connector in the prompt:

```bash
ddn connector init <connector-name> -i
```

#### Step 2.1: Choose `hasura/elasticsearch` from the list

#### Step 2.2: Choose a port for the connector

The CLI will ask for a specific port to run the connector on. Choose a port that is not already in use or use the
default suggested port.

#### Step 2.3: Provide the env vars required for the connector

To configure the connector, the following environment variables need to be set:

| Environment Variable          | Description                                                                                                                                                                | Required | Example Value                                                  |
| ----------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | -------------------------------------------------------------- |
| `ELASTICSEARCH_URL`           | The comma-separated list of Elasticsearch host addresses for connection (Use `local.hasura.dev` instead of `localhost` if your connector is running on your local machine) | Yes      | `https://example.es.gcp.cloud.es.io:9200`                      |
| `ELASTICSEARCH_USERNAME`      | The username for authenticating to the Elasticsearch cluster                                                                                                               | Yes      | `admin`                                                        |
| `ELASTICSEARCH_PASSWORD`      | The password for the Elasticsearch user account                                                                                                                            | Yes      | `default`                                                      |
| `ELASTICSEARCH_API_KEY`       | The Elasticsearch API key for authenticating to the Elasticsearch cluster                                                                                                  | No       | `ABCzYWk0NEI0aDRxxxxxxxxxx1k6LWVQa2gxMUpRTUstbjNwTFIzbGoyUQ==` |
| `ELASTICSEARCH_CA_CERT_PATH`  | The path to the Certificate Authority (CA) certificate for verifying the Elasticsearch server's SSL certificate                                                            | No       | `/etc/connector/cacert.pem`                                    |
| `ELASTICSEARCH_INDEX_PATTERN` | The pattern for matching Elasticsearch indices, potentially including wildcards, used by the connector                                                                     | No       | `hasura*`                                                      |

## Step 3: Introspect the connector

```bash
ddn connector introspect <connector-name>
```

This will generate a `configuration.json` file that will have the schema of your Elasticsearch DB.

## Step 4: Add your resources

```bash
ddn connector-link add-resources <connector-name>
```

This command will track all the indices in your Elasticsearch DB as [Models](https://hasura.io/docs/3.0/supergraph-modeling/models).

## Contributing

Check out our [contributing guide](https://github.com/hasura/ndc-elasticsearch/blob/main/docs/contributing.md) for more details.

## License

The Elasticsearch connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
