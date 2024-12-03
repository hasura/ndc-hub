# DynamoDB for NoSQL Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/latest/connectors/dynamodb/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-dynamodb-blue.svg?style=flat)](https://hasura.io/connectors/dynamodb)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](./readme.md)

With this connector, Hasura allows you to instantly create a real-time GraphQL API on top of your data models in DynamoDB for NoSQL Database containers. This connector supports DynamoDB for NoSQL's functionalities listed in
the table below, allowing for efficient and scalable data operations.

This connector is built using the [Rust Data Connector SDK](https://github.com/hasura/ndc-hub#rusk-sdk) and implements
the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [Connector information in the Hasura Hub](https://hasura.io/connectors/dynamodb)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0)

## Features

Below, you'll find a matrix of all supported features for the DynamoDB for NoSQL connector:

| Feature                         | Supported | Notes |
|---------------------------------|-----------|-------|
| Native Queries + Logical Models | ❌        |       |
| Simple Object Query             | ✅        |       |
| Filter / Search                 | ✅        |       |
| Simple Aggregation              | ❌        |       |
| Sort                            | ✅        |       |
| Paginate                        | ✅        |       |
| Nested Objects                  | ❌        |       |
| Nested Arrays                   | ❌        |       |
| Nested Filtering                | ❌        |       |
| Nested Sorting                  | ❌        |       |
| Nested Relationships            | ❌        |       |

## Documentation

The DynamoDB NoSQL database uses 3 operations to fetch data based on the partition and sort key:

1. GET transaction: Both the partition and sort keys are provided in the filter clause. This is th eleast expensive operation.
2. Query: Partition key is provided in the filter clause.
3. Scan: None of the key attributes are provided. This is the most expensive operation.

Note: To fetch the data using the Global Secondary Index, the collection name should be provided as '<table_name>:<gsi_name>

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes).You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the DynamoDB for NoSQL connector

With the [context set](https://hasura.io/docs/3.0/cli/commands/ddn_context_set/) for an existing subgraph, initialize
the connector by choosing `hasura/dynamodb`. When the wizard runs, you'll also be prompted to enter the following
env vars necessary for your connector to function:

| Name                                      | Description                                                                 | Required |
| ----------------------------------------- | --------------------------------------------------------------------------- | -------- |
| HASURA_DYNAMODB_AWS_ACCESS_KEY_ID         | The AWS DynamoDB access key ID                                              | Yes      |
| HASURA_DYNAMODB_AWS_SECRET_ACCESS_KEY     | The AWS DynamoDB secret access key                                          | Yes      |
| HASURA_DYNAMODB_AWS_REGION                | The region of the AWS DynamoDB database                                     | Yes      |

After the CLI initializes the connector, you'll need to:

- [Introspect](https://hasura.io/docs/3.0/cli/commands/ddn_connector_introspect) the source.
- Add your [models](https://hasura.io/docs/3.0/cli/commands/ddn_model_add),
  [commands](https://hasura.io/docs/3.0/cli/commands/ddn_command_add),
- Create a [new build](https://hasura.io/docs/3.0/cli/commands/ddn_supergraph_build_local).
- Test it by [running your project along with the connector](https://hasura.io/docs/3.0/cli/commands/ddn_run#examples).

## Contributing

We're happy to receive any contributions from the community. Please refer to our
[development guide](https://github.com/hasura/ndc-dynamodb/blob/main/docs/development.md).

## License

The Hasura DynamoDB for NoSQL connector is available under the
[Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
