# Storage Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-storage-blue.svg?style=flat)](https://hasura.io/connectors/storage)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://www.apache.org/licenses/LICENSE-2.0)

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your
cloud storage objects. This connector supports cloud storage functionalities to manage your files on cloud storage, allowing for efficient
and scalable data operations.

This connector is built using the [Go Data Connector SDK](https://github.com/hasura/ndc-sdk-go) and implements the
[Data Connector Spec](https://github.com/hasura/ndc-spec).

- [See the listing in the Hasura Hub](https://hasura.io/connectors/storage)
- [Hasura DDN Documentation](https://hasura.io/docs/3.0)
- [Hasura DDN Quickstart](https://hasura.io/docs/3.0/getting-started/quickstart)

Docs for the Storage data connector:

- [Configuration](https://github.com/hasura/ndc-storage/blob/main/docs/configuration.md)
- [Manage Objects](https://github.com/hasura/ndc-storage/blob/main/docs/objects.md)
- [Security](https://github.com/hasura/ndc-storage/blob/main/SECURITY.md)

## Features

### Supported storage services

| Service              | Supported |
| -------------------- | --------- |
| AWS S3               | ✅ (\*)   |
| Google Cloud Storage | ✅        |
| Azure Blob Storage   | ✅        |
| MinIO                | ✅ (\*)   |
| Cloudflare R2        | ✅ (\*)   |
| DigitalOcean Spaces  | ✅ (\*)   |

(\*): Support Amazon S3 Compatible Cloud Storage providers. The connector uses [MinIO Go Client SDK](https://github.com/minio/minio-go) behind the scenes.

### Supported features

Below, you'll find a matrix of all supported features for the Storage connector:

| Feature                | Supported | Notes |
| ---------------------- | --------- | ----- |
| List Buckets           | ✅        |       |
| Create Bucket          | ✅        |       |
| Update Bucket          | ✅        |       |
| Delete Bucket          | ✅        |       |
| List Objects           | ✅        |       |
| Upload Object          | ✅        |       |
| Download Object        | ✅        |       |
| Delete Object          | ✅        |       |
| Generate Presigned-URL | ✅        |       |

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)
5. Authentication credentials of cloud storage services.

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes).You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the Storage connector

With the [context set](https://hasura.io/docs/3.0/cli/commands/ddn_context_set/) for an existing subgraph, initialize
the connector:

```sh
ddn connector init -i
```

When the wizard runs, choose `hasura/storage` connector. AWS S3 environment variables are the default settings in the interactive prompt. You'll be prompted to enter the following env vars necessary for your connector to function. If you want to use other storage providers you need to manually configure the `configuration.yaml` file and add the required environment variable mappings to the subgraph definition.

After the CLI initializes the connector, you'll need to:

- [Introspect](https://hasura.io/docs/3.0/cli/commands/ddn_connector_introspect) the source.
- Add your [models](https://hasura.io/docs/3.0/cli/commands/ddn_model_add),
  [commands](https://hasura.io/docs/3.0/cli/commands/ddn_command_add), and
  [relationships](https://hasura.io/docs/3.0/cli/commands/ddn_relationship_add).
- Create a [new build](https://hasura.io/docs/3.0/cli/commands/ddn_supergraph_build_local).
- Test it by [running your project along with the connector](https://hasura.io/docs/3.0/cli/commands/ddn_run#examples).

## License

The Storage connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
