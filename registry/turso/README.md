## Turso Connector

TODO: Update README!

The Turso Data Connector allows for connecting to a libSQL database. This uses the [Typescript Data Connector SDK](https://github.com/hasura/ndc-sdk-typescript) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec). 

In order to use this connector you will need a Turso database setup. This connector currently only supports querying. 

## Before you get started
It is recommended that you:

* Clone this repo and run npm install
* Install the [Hasura3 CLI](https://github.com/hasura/v3-cli#hasura-v3-cli)
* Log in via the CLI
* Install the [connector plugin](https://hasura.io/docs/latest/hasura-cli/connector-plugin/)
* Install [VSCode](https://code.visualstudio.com)
* Install the [Hasura VSCode Extension](https://marketplace.visualstudio.com/items?itemName=HasuraHQ.hasura)
* (Optionally) Setup a [Turso Database instance](https://docs.turso.tech/tutorials/get-started-turso-cli)

## Deployment for Hasura Users
To deploy a connector and use it in a Hasura V3 project, follow these steps:

1. Create a Hasura V3 project (or use an existing project)

2. Generate a configuration file for your Turso Database.
    Start the configuration server by running `npm run start-server` and then send a request with your database credentials.
    You can see example http requests in the http_requests folder. You may need to edit your credentials here. You'll want to place the output in a JSON file named config.json
3. Once you have a configuration file, you can deploy the connector onto Hasura Cloud

Ensure you are logged in to Hasura CLI

```hasura3 cloud login --pat 'YOUR-HASURA-TOKEN'```

From there, you can deploy the connector:

```hasura3 connector create turso:v1 --github-repo-url https://github.com/hasura/ndc-turso/tree/main --config-file ./config.json```


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