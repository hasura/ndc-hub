# Storage Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-storage-blue.svg?style=flat)](https://hasura.io/connectors/storage)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://www.apache.org/licenses/LICENSE-2.0)

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your
cloud storage objects. This connector supports cloud storage functionalities to manage your files on cloud storage, allowing for efficient
and scalable data operations.

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

| Service                  | Type     | List Buckets | Create Bucket | Update Bucket | Delete Bucket | List Objects | Upload | Download | Delete Object | Soft-Delete | Presigned-URL |
| ------------------------ | -------- | ------------ | ------------- | ------------- | ------------- | ------------ | ------ | -------- | ------------- | ----------- | ------------- |
| AWS S3 (\*)              | `s3`     | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ❌          | ✅            |
| Google Cloud Storage     | `gcs`    | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ✅          | ✅            |
| Azure Blob Storage       | `azblob` | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ✅          | ✅            |
| File System              | `fs`     | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ❌          | ❌            |
| MinIO (\*)               | `s3`     | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ❌          | ✅            |
| Cloudflare R2 (\*)       | `s3`     | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ❌          | ✅            |
| DigitalOcean Spaces (\*) | `s3`     | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ❌          | ✅            |

(\*): Support Amazon S3 Compatible Cloud Storage providers. The connector uses [MinIO Go Client SDK](https://github.com/minio/minio-go) behind the scenes.

## Build on Hasura DDN

[Get started](https://hasura.io/docs/3.0/how-to-build-with-ddn/with-storage) by connecting your preferred cloud storage
provider to a Hasura DDN project.

## Fork the connector

You can fork the [connector's repo](https://github.com/hasura/ndc-storage) and iterate on it yourself.

## License

The Storage connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
