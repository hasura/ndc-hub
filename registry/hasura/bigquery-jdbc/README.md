# BigQuery Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/getting-started/overview/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-bigquery-blue.svg?style=flat)](https://hasura.io/connectors/bigquery-jdbc)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](./readme.md)

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your data models in
BigQuery. This connector supports BigQuery's functionalities listed in the table below, allowing for efficient and
scalable data operations. Additionally, users benefit from all the powerful features of Hasura’s Data Delivery Network
(DDN) platform, including query pushdown capabilities that delegate query operations to the database, thereby enhancing
query optimization and performance.

This connector implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [Connector information in the Hasura Hub](https://hasura.io/connectors/bigquery-jdbc)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0)

## Features

Below, you'll find a matrix of all supported features for the BigQuery connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models | ❌        |       |
| Native Mutations                | ❌        |       |
| Simple Object Query             | ✅        |       |
| Filter / Search                 | ✅        |       |
| Simple Aggregation              | ✅        |       |
| Sort                            | ✅        |       |
| Paginate                        | ✅        |       |
| Table Relationships             | ❌        |       |
| Views                           | ✅        |       |
| Remote Relationships            | ✅        |       |
| Custom Fields                   | ❌        |       |
| Mutations                       | ❌        |       |
| Distinct                        | ❌        |       |
| Enums                           | ❌        |       |
| Naming Conventions              | ❌        |       |
| Default Values                  | ❌        |       |
| User-defined Functions          | ❌        |       |

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes).You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the BigQuery connector

With the [context set](https://hasura.io/docs/3.0/cli/commands/ddn_context_set/) for an existing subgraph, initialize
the connector:

```sh
ddn connector init -i
```

When the wizard runs, you'll be prompted to enter the following env vars necessary for your connector to function:

| Name         | Description                                                  | Required |
| ------------ | ------------------------------------------------------------ | -------- |
| JDBC_URL     | The JDBC URL to connect to the database                      | Yes      |

After the CLI initializes the connector, you'll need to:

### Configuring your JDBC connection string
The official BigQuery JDBC driver is used. You can find documentation on configuring the JDBC connection string
[here](https://cloud.google.com/bigquery/docs/reference/odbc-jdbc-drivers#current_jdbc_driver). As an example using a service account with a full key file downloaded from google:
```
APP_FOO_JDBC_URL=jdbc:bigquery://https://www.googleapis.com/bigquery/v2:443;ProjectId=project-id;DefaultDataset=dataset;OAuthType=0;OAuthServiceAcctEmail=service-account-email;OAuthPvtKey=/etc/connector/key.json;
```
**Note:** `ProjectId` and `DefaultDataset` are required in your JDBC connection string.
**Note:** since the files get mounted in docker it is import the file path is `/etc/connector/<your-key-file>.json`

Make sure you place you `key.json` in the connector folder `/<subgraph>/connector/<connectorname>/key.json`. The key
should be the full key downloaded from google cloud console that looks like:
```
{
  "type": "service_account",
  "project_id": "project-id",
  "private_key_id": "private-key-id",
  "private_key": "-----BEGIN PRIVATE KEY-----\nprivate-key\n-----END PRIVATE KEY-----\n",
  "client_email": "service-account-email",
  "client_id": "client-id",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/service-account-email"
}
```

Once that is done you'll need to:

- [Introspect](https://hasura.io/docs/3.0/cli/commands/ddn_connector_introspect) the source.
- Add your [models](https://hasura.io/docs/3.0/cli/commands/ddn_model_add),
  [commands](https://hasura.io/docs/3.0/cli/commands/ddn_command_add), and
  [relationships](https://hasura.io/docs/3.0/cli/commands/ddn_relationship_add).
- Create a [new build](https://hasura.io/docs/3.0/cli/commands/ddn_supergraph_build_local).
- Test it by [running your project along with the connector](https://hasura.io/docs/3.0/cli/commands/ddn_run#examples).

## License

The Hasura BigQuery connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
