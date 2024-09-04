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

## Before you get Started

1. The [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and [Docker](https://docs.docker.com/engine/install/) installed
2. A [supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
3. A [subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

The steps below explain how to Initialize and configure a connector for local development. You can learn how to deploy a connector — after it's been configured — [here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the Python connector

### Step 1: Authenticate your CLI session

```bash
ddn auth login
```

### Step 2: Configure the connector

Once you have an initialized supergraph and subgraph, run the initialization command in interactive mode while providing a name for the connector in the prompt:

```bash
ddn connector init python -i
```

#### Step 2.1: Choose the `hasura/python` option from the list

#### Step 2.2: Choose a port for the connector

The CLI will ask for a specific port to run the connector on. Choose a port that is not already in use or use the default suggested port.

### Step 3: Introspect the connector

Introspecting the connector will generate a `config.json` file and a `python.hml` file.

```bash
ddn connector introspect python
```

### Step 4: Add your resources

You can add the models, commands, and relationships to your API by tracking them which generates the HML files. 

```bash
ddn connector-link add-resources python
```

### Step 5: Run your connector

You can run your connector locally, or include it in the docker setup.

#### Run the connector in Docker

To include your connector in the docker setup, include its compose file at the top of your supergraph `compose.yaml` file like this:

```yaml
include:
  - path: app/connector/python/compose.yaml
```

#### Run the connector locally

To run your connector outside of Docker first go into the connector directory:

`cd app/connector/python`

Install the requirements:

`pip3 install -r requirements.txt`

Then run the connector locally:

```ddn connector setenv --connector connector.yaml -- python3 functions.py serve```

## Documentation

View the full documentation for the Python Lambda connector [here](https://github.com/hasura/ndc-python-lambda/blob/main/docs/index.md).

## Contributing

Check out our [contributing guide](https://github.com/hasura/ndc-python-lambda/blob/main/docs/contributing.md) for more details.

## License

The Turso connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).