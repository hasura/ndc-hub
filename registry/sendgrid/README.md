## Overview

This connector uses the SendGrid v3 API to:

* List email templates
* Send emails

https://github.com/hasura/ndc-sendgrid/tree/main#sendgrid-connector

* [Create a SendGrid API account](https://signup.sendgrid.com/)
* [Create an API key](https://app.sendgrid.com/settings/api_keys)
* Create a share service token

You will need the Hasura
[V3 CLI](https://github.com/hasura/v3-cli)
and
[Connector Plugin](https://hasura.io/docs/latest/hasura-cli/connector-plugin/)
installed to use this connector.


## Deployment

You will need to have a configuration file available with your sendgrid credentials in the following format:

```
> cat sendgrid.connector.configuration.json
{"version": 1, "sendgrid_api_key": "YOUR-API-KEY-HERE" }
```

Deploy and name the connector with the following command referencing your config:

> hasura3 connector create sendgrid:v1 --github-repo-url https://github.com/hasura/ndc-sendgrid/tree/main --volume ./sendgrid.connector.configuration.json:/config.json --env SERVICE_TOKEN_SECRET=MY-SERVICE-TOKEN

Monitor the deployment status by name:

> hasura connector status sendgrid:v1

List your connector with its deployed URL:

> hasura connector list

```
sendgrid:v1 https://connector-9XXX7-hyc5v23h6a-ue.a.run.app active
```


## Usage

Include the connector URL in your Hasura V3 project metadata:

```json
[
  {
      "kind": "dataSource",
      "name": "sendgrid",
      "dataConnectorUrl": "https://connector-9XXX7-hyc5v23h6a-ue.a.run.app",
      "schema": {}
  }
  ...
]
```

## Troubleshooting

Please [https://github.com/hasura/ndc-sendgrid/issues/new](submit a Github issue)
if you encounter any problems!

