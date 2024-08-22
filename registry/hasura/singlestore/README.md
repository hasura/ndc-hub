# SingleStore Data Connector

![Logo Image](docs/singlestore_logo_horizontal_color_on-white_rgb.png)

<!-- TODO: update when connector will be published -->
[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/latest/connectors/singesltore/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-singlestore-blue.svg?style=flat)](https://hasura.io/connectors/singlestore)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](./readme.md)

The Hasura SingleStore Connector ("the connector") enables you to connect to a SingleStore database and gives instant access to a GraphQL API on top of your data.

This connector is built using the [Typescript Data Connector SDK](https://github.com/hasura/ndc-sdk-typescript) and, it implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

<!-- TODO: update when connector will be published -->
- [See the listing in the Hasura Hub](https://hasura.io/connectors/singlestore)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0)

## Features

The following matrix lists the features supported by the Hasura SingleStore connector:

| Feature                         | Supported | Notes |
| ------------------------------- | --------- | ----- |
| Native Queries + Logical Models | ❌     |       |
| Simple Object Query             | ✅     |       |
| Filter / Search                 | ✅     |       |
| Simple Aggregation              | ✅     |       |
| Sort                            | ✅     |       |
| Paginate                        | ✅     |       |
| Table Relationships             | ✅     |       |
| Views                           | ✅     |       |
| Distinct                        | ✅     |       |
| Remote Relationships            | ✅     |       |
| Mutations                       | ❌     | coming soon       |

## Prerequisites

Ensure that the following prerequisites are met before using the connector:

1. Install [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and [Docker](https://docs.docker.com/engine/install/).
2. Initialize a [supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph).
3. Initialize a [subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph).
4. An active [SingleStore](https://www.singlestore.com/) deployment that serves as the data source for the API.


## Using the SingleStore connector

The following steps explain how to initialize and configure the connector for local development. For information on deploying a connector after it has been configured, refer to [Deploy a Connector](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).
<!-- TODO: test steps after the connector will be published to hub -->
### Step 1: Authenticate your CLI session

```bash
ddn auth login
```

### Step 2: Configure the connector

After the supergraph and subgraph have been initialized (refer to [Prerequisites](#Prerequisites), run the initialization command in interactive mode. Specify a name for the connector in the command.

```bash
ddn connector init <connector-name>  -i
```

#### Step 2.1: Select the SingleStore connector

From the list of connectors, select `singlestore`.

#### Step 2.2: Select a port for the connector
The CLI prompts you to specify a port for running the connector. Either select the default port suggested or choose a port that is not currently in use.

#### Step 2.3: Provide the environment variable(s) for the connector

Specify the connection string used to connect to the SingleStore database using the `SINGLESTORE_URL` variable. The connection string must be in the `mysql://[<username>[:<password>]][@<host>:[<port>]]/[<database>][?<key1>=<value1>[&<key2>=<value2>]]` format. 

For example:

```env
SINGLESTORE_URL=singlestore://username:password@hostname/database
```
The connector uses [MySQL2](https://sidorares.github.io/node-mysql2/docs) library to establish a connection.
For more information, refer to [Connection options](https://www.npmjs.com/package/mysql#connection-options) and [Pool options](https://www.npmjs.com/package/mysql#pool-options).

Alternatively, you can also set following environment variables instead of the `SINGLELSTORE_URL` variable:

| Name                                | Default     | Description |
|-------------------------------------|-------------|-------------|
| SINGLESTORE_HOST                    | localhost   | Hostname of the SingleStore database to connect with. |
| SINGLESTORE_PORT                    | 3306        | Port number of the SingleStore database. |
| SINGLESTORE_PASSWORD                |             | Password of the SingleStore database user. |
| SINGLESTORE_DATABASE                |             | Name of the SingleStore database to connect with. |
| SINGLESTORE_SSL_CA                  |             | Path to the trusted CA certificate file. |
| SINGLESTORE_SSL_CERT                |             | Path to the certificate chain file in PEM format. |
| SINGLESTORE_SSL_KEY                 |             | Path to the private key file in PEM format. |
| SINGLESTORE_SSL_CIPHERS             |             | Cipher suite specification. If specified, it replaces the default value. |
| SINGLESTORE_SSL_PASSPHRASE          |             | Cipher suite specification. If specified, it replaces the default value. |
| SINGLESTORE_SSL_REJECT_UNAUTHORIZED | true        | If enabled, the server rejects any connection that is not authorized with the list of supplied CAs. |

If `SINGLELSTORE_URL` is set these variables are ignored.

### Step 3: Introspect your data source

After configuring the connector, use the CLI to introspect the SingleStore database and create a source-specific configuration file for the connector (`configuration.json`).

```bash
ddn connector introspect <connector-name>
```

## Step 4: Add your resources

This command will create `.hml` files for each table and view in your database 
and an `.hml` file with information about SingleStore data types.

```bash
ddn connector-link add-resources <connector-name>
```

## Step 5. Add relationships

SingleStore does not support foreign keys. Relationships between tables must be added manually. You can define relationships by appending relationship information to the `.hml` files generated in the previous step.
For information on defining relationships, refer to [Relationships](https://hasura.io/docs/3.0/supergraph-modeling/relationships/).
For example, to add a relationship from a `message` table to the `user` table, append following text to the 
`DbMessage.hml` file:
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
