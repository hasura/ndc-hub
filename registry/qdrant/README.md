## Qdrant Connector Overview

The Qdrant Data Connector allows for connecting to a Qdrant instance giving you an instant GraphQL API that supports querying on top of your data. This uses the [Typescript Data Connector SDK](https://github.com/hasura/ndc-sdk-typescript) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec). 

In order to use this connector you will need a Qdrant database setup. This connector currently only supports querying. 

## Before you get started

It is recommended that you:

* Setup a [Qdrant Database instance](https://qdrant.tech/)
* Install the [Hasura3 CLI](https://github.com/hasura/v3-cli#hasura-v3-cli)
* Log in via the CLI
* Install the [connector plugin](https://hasura.io/docs/latest/hasura-cli/connector-plugin/)
* Install [VSCode](https://code.visualstudio.com)
* Install the [Hasura VSCode Extension](https://marketplace.visualstudio.com/items?itemName=HasuraHQ.hasura)

## Deployment For Hasura Users

To deploy a connector and use it in a Hasura V3 project, follow these steps:

1. Create a Hasura V3 project (or use an existing project)

2. Generate a configuration file for your Qdrant Database, there are 2 ways to get the configuration file.

    First you'll need to clone [this repo](https://github.com/hasura/ndc-qdrant), and run ```npm install```
    i. The easiest way to generate a configuration file is to run the generate-config script using ts-node. 

    When running this script specify:

    --url The URL where Qdrant is hosted

    --key The API key for connecting to the Qdrant Client.

    --output The name of the file to store the configuration in

    Example Usage:

    ```ts-node generate-config --url https://qdrant-url --key QdrantApiKey --output config.json```
    
    ii. You can also run the connector in configuration mode and generate the config file using CURL.

    ```ts-node ./src/index.ts configuration serve```

    You can then send a CURL request specifying the qdrant_url and qdrant_api_key to get the configuration file.

    Example:

    ```curl -X POST -H "Content-Type: application/json" -d '{"qdrant_url": "https://link-to-qdrant.cloud.qdrant.io", "qdrant_api_key": "QdrantApiKey"}' http://localhost:9100 > config.json```

3. Once you have a configuration file, you can deploy the connector onto Hasura Cloud

Ensure you are logged in to Hasura CLI

```hasura3 cloud login --pat 'YOUR-HASURA-TOKEN'```

From there, you can deploy the connector:

```hasura3 connector create qdrant:v1 --github-repo-url https://github.com/hasura/ndc-qdrant/tree/main --config-file ./config.json```

## Usage

Once your connector is deployed, you can get the URL of the connector using:
```hasura3 connector list```

```
my-cool-connector:v1 https://connector-9XXX7-hyc5v23h6a-ue.a.run.app active
```

In order to use the connector once deployed you will first want to reference the connector in your project metadata:

```yaml
kind: "AuthConfig"
allowRoleEmulationFor: "admin"
webhook:
  mode: "POST"
  webhookUrl: "https://auth.pro.hasura.io/webhook/ddn?role=admin"
---
kind: DataConnector
version: v1
definition:
  name: my_connector
  url:
    singleUrl: 'https://connector-9XXX7-hyc5v23h6a-ue.a.run.app'
```

If you have the [Hasura VSCode Extension](https://marketplace.visualstudio.com/items?itemName=HasuraHQ.hasura) installed
you can run the following code actions:

* `Hasura: Refresh data source`
* `Hasura: Track all collections / functions ...`

This will integrate your connector into your Hasura project which can then be deployed or updated using the Hasura3 CLI:

```
hasura3 cloud build create --project-id my-project-id --metadata-file metadata.hml
```

## Service Authentication

If you don't wish to have your connector publically accessible then you must set a service token by specifying the  `SERVICE_TOKEN_SECRET` environment variable when creating your connector:

* `--env SERVICE_TOKEN_SECRET=SUPER_SECRET_TOKEN_XXX123`

Your Hasura project metadata must then set a matching bearer token:

```yaml
kind: DataConnector
version: v1
definition:
  name: my_connector
  url:
    singleUrl: 'https://connector-9XXX7-hyc5v23h6a-ue.a.run.app'
  headers:
    Authorization:
      value: "Bearer SUPER_SECRET_TOKEN_XXX123"
```

While you can specify the token inline as above, it is recommended to use the Hasura secrets functionality for this purpose:

```yaml
kind: DataConnector
version: v1
definition:
  name: my_connector
  url:
    singleUrl: 'https://connector-9XXX7-hyc5v23h6a-ue.a.run.app'
  headers:
    Authorization:
      valueFromSecret: BEARER_TOKEN_SECRET
```

NOTE: This secret should contain the `Bearer ` prefix.


## Default Collection Parameters:

You'll find that each collection on your graph is parameterized, and that you have the ability to pass in the following parameters as collection arguments:

vector
positive
negative

These will allow you to perform vector searches, or to get recommendations.

You can pass in a search vector to the vector parameter, which is a flat list of floats. This will typically be the output from some embedding model, and it will return results ordered by closest match. You'll likely want to ensure that you are passing a limit on all your queries.

You can also pass in an array of ID's to the positive and negative parameters to provide example data-points. This is an easy way to get recommendations without having to manage or deal with passing around entire vectors. If you know the ID of some positive and negative data-points, you can simply pass the ID's. You must provide at least 1 positive example when using this. You can provide a list of positive examples, a list of positive and a list of negative, but you cannot provide only a list of negative examples.

You can read more about these parameters [here](https://qdrant.tech/documentation/concepts/search/)