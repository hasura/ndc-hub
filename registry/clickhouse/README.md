## Overview

The Clickhouse Native Data Connector allows for connecting to a Clickhouse instance giving you an instant GraphQL API on top of your Clickhouse data.
This uses the [Rust Data Connector SDK](https://github.com/hasura/ndc-hub#rusk-sdk) from the [Data connector Hub](https://github.com/hasura/ndc-hub) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

* [Clickhouse Connector information in the Hasura Connectors directory](https://hasura.io/connectors/clickhouse)
* TODO: Docs Link

In order to use this connector you will need to:

* Create a [Clickhouse account](https://clickhouse.cloud/signUp?loc=nav-get-started)
* Log in to A Hasura CLI Session
* Create a Pre-Shared Token for service authentication between the Hasura V3 Engine and your connector

## Features

TODO

## Deployment

The following steps will allow you to deploy the connector and use it in a Hasura V3 project:

* Create a Hasura V3 Project (or use an existing project)
* Ensure that you have a metadata definition
* Create a configuration for the Clickhouse Connector referencing your credentials:
     `clickhouse.connector.configuration.json`
     You have 2 options for the config:
     1. The easiest option is to is to run the connector locally in config mode:
     ```
     cargo run -- configuration serve --port 5000
     curl -X POST -d '{"connection": {"username": "your_username", "password": "your_password", "url": "your_clickhouse_url"}, "tables": []}' http://localhost:5000 > clickhouse.connector.configuration.json
     ```
     2. The other option is to manually write your config that follows this pattern:
     ```
     {
        "connection": {
          "username": "your_username",
          "password": "your_password",
          "url": "your_clickhouse_url"
        },
        "tables": [
          {
            "name": "TableName",
            "schema": "SchemaName",
            "alias": "TableAlias",
            "primary_key": { "name": "TableId", "columns": ["TableId"] },
            "columns": [
              { "name": "TableId", "alias": "TableId", "data_type": "Int32" },
            ]
          }
        ]
      }
     ```
* Run the following command to deploy the connector
* Ensure you are logged in to Hasura CLI
     ```
     hasura3 cloud login --pat 'YOUR-HASURA-TOKEN'
     ```
* Deploy the connector
     ```
     hasura3 connector create clickhouse:v1 \
     --github-repo-url https://github.com/hasura/clickhouse_ndc/tree/main \
     --config-file ./clickhouse.connector.configuration.json
     ```
* Ensure that your deployed connector is referenced from your metadata with the service token
* Edit your metadata using the LSP support to import the defined schema, functions, procedures
* Deploy or update your Hasura cloud project
     ```
     hasura3 cloud build create --project-id my-project-id \
     --metadata-file metadata.json my-build-id
     ```
* View in your cloud console, access via the graphql API
