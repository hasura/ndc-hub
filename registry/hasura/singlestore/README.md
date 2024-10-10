# SingleStore Data Connector

<a href="https://www.singlestore.com/"><img src="https://github.com/singlestore-labs/singlestore-hasura-connector/blob/main/docs/logo_primary_singlestore_black.png" align="right" width="200"></a>

<!-- TODO: update when connector will be published -->

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/latest/connectors/singesltore/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-singlestore-blue.svg?style=flat)](https://hasura.io/connectors/singlestore)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](./readme.md)

The Hasura SingleStore Connector ("the connector") enables you to connect to a SingleStore database and gives instant
access to a GraphQL API on top of your data.

This connector is built using the [Typescript Data Connector SDK](https://github.com/hasura/ndc-sdk-typescript) and, it
implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

<!-- TODO: update when connector will be published -->

- [See the listing in the Hasura Hub](https://hasura.io/connectors/singlestore)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0)

## Features

The following matrix lists the features supported by the Hasura SingleStore connector:

| Feature                         | Supported | Notes       |
| ------------------------------- | --------- | ----------- |
| Native Queries + Logical Models | ❌        |             |
| Simple Object Query             | ✅        |             |
| Filter / Search                 | ✅        |             |
| Simple Aggregation              | ✅        |             |
| Sort                            | ✅        |             |
| Paginate                        | ✅        |             |
| Table Relationships             | ✅        |             |
| Views                           | ✅        |             |
| Distinct                        | ✅        |             |
| Remote Relationships            | ✅        |             |
| Mutations                       | ❌        | coming soon |

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)
5. An active [SingleStore](https://www.singlestore.com/) deployment that serves as the data source for the API.

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes).You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the SingleStore connector

With the [context set](https://hasura.io/docs/3.0/cli/commands/ddn_context_set/) for an existing subgraph, initialize
the connector:

```sh
ddn connector init -i
```

When the wizard runs, you'll be prompted to enter the following env vars necessary for your connector to function:

| Name                                | Default   | Description                                                                                         |
| ----------------------------------- | --------- | --------------------------------------------------------------------------------------------------- |
| SINGLESTORE_HOST                    | localhost | Hostname of the SingleStore database to connect with.                                               |
| SINGLESTORE_PORT                    | 3306      | Port number of the SingleStore database.                                                            |
| SINGLESTORE_USER                    |           | SingleStore user to authenticate as.                                                                |
| SINGLESTORE_PASSWORD                |           | Password of the SingleStore database user.                                                          |
| SINGLESTORE_DATABASE                |           | Name of the SingleStore database to connect with.                                                   |
| SINGLESTORE_SSL_CA                  |           | CA certificate.                                                                                     |
| SINGLESTORE_SSL_CERT                |           | Certificate chain in PEM format.                                                                    |
| SINGLESTORE_SSL_KEY                 |           | Private key in PEM format.                                                                          |
| SINGLESTORE_SSL_CIPHERS             |           | Cipher suite specification. If specified, it replaces the default value.                            |
| SINGLESTORE_SSL_PASSPHRASE          |           | Shared passphrase used for a single private key.                                                    |
| SINGLESTORE_SSL_REJECT_UNAUTHORIZED | true      | If enabled, the server rejects any connection that is not authorized with the list of supplied CAs. |

The connector uses [MySQL2](https://sidorares.github.io/node-mysql2/docs) library to establish a connection. For more
information, refer to [Connection options](https://www.npmjs.com/package/mysql#connection-options) and
[Pool options](https://www.npmjs.com/package/mysql#pool-options).

After the CLI initializes the connector, you'll need to:

- [Introspect](https://hasura.io/docs/3.0/cli/commands/ddn_connector_introspect) the source.
- Add your [models](https://hasura.io/docs/3.0/cli/commands/ddn_model_add),
  [commands](https://hasura.io/docs/3.0/cli/commands/ddn_command_add), and
  [relationships](https://hasura.io/docs/3.0/cli/commands/ddn_relationship_add).
- Create a [new build](https://hasura.io/docs/3.0/cli/commands/ddn_supergraph_build_local).
- Test it by [running your project along with the connector](https://hasura.io/docs/3.0/cli/commands/ddn_run#examples).

### Note

SingleStore does not support foreign keys. Relationships between tables must be added manually. You can define
relationships by appending relationship information to the `.hml` files generated in the previous step. For information
on defining relationships, refer to [Relationships](https://hasura.io/docs/3.0/supergraph-modeling/relationships/). For
example, to add a relationship from a `message` table to the `user` table, append following text to the `DbMessage.hml`
file:

```hml
---
kind: Relationship
version: v1
definition:
  name: user
  sourceType: DbMessage
  target:
    model:
      name: DbUser
      subgraph: app
      relationshipType: Object
  mapping:
    - source:
        fieldPath:
          - fieldName: userId
      target:
        modelField:
          - fieldName: id
  description: The user details for a message
```

## License

The SingleStore connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
