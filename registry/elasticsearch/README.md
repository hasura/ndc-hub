# Elasticsearch Connector

The Hasura Elasticsearch Connector allows for connecting to a Elasticsearch search engine, giving you an instant
GraphQL API on top of your Elasticsearch data.

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
| Nested Filtering                | ✅         |       |
| Nested Sorting                  | ✅         |       |
| Nested Relationships            | ❌         |       |


## Before you get Started

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Install the [CLI](https://hasura.io/docs/3.0/cli/installation/)
3. [Create a project](https://hasura.io/docs/3.0/getting-started/create-a-project)

## Using the connector

To use the Elasticsearch connector, follow these steps in a Hasura project:

1. Add the connector:

   ```bash
   ddn add connector-manifest es_connector --subgraph app --hub-connector hasura/elasticsearch --type cloud
   ```

   In the snippet above, we've used the subgraph `app` as it's available by default; however, you can change this
   value to match any [subgraph](https://hasura.io/docs/3.0/project-configuration/subgraphs) which you've created in your project.

2. Add your Elasticsearch credentials:

   Open your project in your text editor and open the `base.env.yaml` file in the root of your project. Then, add
   `ES_CONNECTOR_URL`, `ES_CONNECTOR_USERNAME` and `ES_CONNECTOR_PASSWORD` environment variables under the `app` subgraph:

   ```yaml
   supergraph: {}
   subgraphs:
     app:
       ES_CONNECTOR_URL: "<YOUR_ELASTICSEARCH_URL>"
       ES_CONNECTOR_USERNAME: "<YOUR_ELASTICSEARCH_USERNAME>"
       ES_CONNECTOR_PASSWORD: "<YOUR_ELASTICSEARCH_PASSWORD>"
   ```

   Next, update your `/app/es_connector/connector/es_connector.build.hml` file to reference this new environment
   variable:

   ```yaml
   # other configuration above
   ELASTICSEARCH_URL:
     valueFromEnv: ES_CONNECTOR_URL
   ELASTICSEARCH_USERNAME:
     valueFromEnv: ES_CONNECTOR_USERNAME
   ELASTICSEARCH_PASSWORD:
     valueFromEnv: ES_CONNECTOR_PASSWORD
   ```

   Notice, when we use an environment variable, we must change the key to `valueFromEnv` instead of `value`. This tells
   Hasura DDN to look for the value in the environment variable we've defined instead of using the value directly.

3. Update the connector manifest and the connector link

   These two steps will (1) allow Hasura to introspect your data source and complete the configuration and (2) deploy the
   connector to Hasura DDN:

   ```bash
   ddn update connector-manifest es_connector
   ```

   ```bash
   ddn update connector-link es_connector
   ```

4. Add Environment Variables

To configure the connector, the following environment variables need to be set:

| Environment Variable          | Description                                                                                                     | Required | Example Value                                                  |
| ----------------------------- | --------------------------------------------------------------------------------------------------------------- | -------- | -------------------------------------------------------------- |
| `ELASTICSEARCH_URL`           | The comma-separated list of Elasticsearch host addresses for connection                                         | Yes      | `https://example.es.gcp.cloud.es.io:9200`                      |
| `ELASTICSEARCH_USERNAME`      | The username for authenticating to the Elasticsearch cluster                                                    | Yes      | `admin`                                                        |
| `ELASTICSEARCH_PASSWORD`      | The password for the Elasticsearch user account                                                                 | Yes      | `default`                                                      |
| `ELASTICSEARCH_API_KEY`       | The Elasticsearch API key for authenticating to the Elasticsearch cluster                                       | No       | `ABCzYWk0NEI0aDRxxxxxxxxxx1k6LWVQa2gxMUpRTUstbjNwTFIzbGoyUQ==` |
| `ELASTICSEARCH_CA_CERT_PATH`  | The path to the Certificate Authority (CA) certificate for verifying the Elasticsearch server's SSL certificate | No       | `/etc/connector/cacert.pem`                                    |
| `ELASTICSEARCH_INDEX_PATTERN` | The pattern for matching Elasticsearch indices, potentially including wildcards, used by the connector          | No       | `hasura*`                                                      |
