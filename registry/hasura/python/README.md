# Hasura Python Lambda Connector

<a href="https://www.python.org/"><img src="https://github.com/hasura/ndc-python-lambda/blob/main/docs/logo.svg" align="right" width="200"></a>

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/connectors/python)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-python-blue.svg?style=flat)](https://hasura.io/connectors/python)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](https://github.com/hasura/ndc-python-lambda/blob/main/LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](https://github.com/hasura/ndc-python-lambda/blob/main/README.md)

This connector allows you to write Python code and call it using Hasura!

With Hasura, you can integrate -- and even host -- this business logic directly with Hasura DDN and your API.

You can handle custom business logic using the Python Lambda data connector. Using this connector, you can transform or enrich data before it reaches your customers, or perform any other business logic you may need.

You can then integrate these functions as individual commands in your metadata and API. This process enables you to simplify client applications and speed up your backend development!

This connector is built using the [Python Data Connector SDK](https://github.com/hasura/ndc-sdk-python) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

The steps below explain how to initialize and configure a connector for local development. You can learn how to deploy a
connector — after it's been configured — [here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the Python Lambda connector

Check out the [Hasura docs here](https://hasura.io/docs/3.0/business-logic/python#add-the-python-connector-to-a-project) to get started with the Python Lambda connector.

## Documentation

View the full documentation for the Python Lambda connector [here](https://github.com/hasura/ndc-python-lambda/blob/main/docs/index.md).

## Contributing

Check out our [contributing guide](https://github.com/hasura/ndc-python-lambda/blob/main/docs/contributing.md) for more details.

## License

The Turso connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).

