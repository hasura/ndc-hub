## Overview

The Hasura PostgreSQL Connector allows for connecting Hasura to a PostgreSQL database giving you an instant GraphQL API on top of your PostgreSQL data.

As much as possible we attempt to provide explicit support for database projects that identify as being derived from PostgreSQL such as [AWS RDS and Aurora PostgreSQL](https://aws.amazon.com/rds/aurora/).

Data Connectors are the way to connect the Hasura Data Delivery Network (DDN) to external data sources. A data connector is an HTTP service that exposes a set of APIs that Hasura uses to communicate with the data source. Data connectors are built to conform to the [NDC Specification](https://hasura.github.io/ndc-spec/overview.html) using one of Hasura's available SDKs. The data connector is responsible for interpreting work to be done on behalf of the Hasura Engine, using the native query language of the data source.

The `ndc-postgres` data connector is open source and can be found in the [ndc-postgres GitHub repository](https://github.com/hasura/ndc-postgres).

Visit the
[Hasura DDN PostgreSQL Documentation](https://hasura.io/docs/3.0/connectors/postgresql/)
for more information about specific features that are available for the PostgreSQL Connector.

## Deployment

The connector is hosted by Hasura and can be used from the [Hasura v3 Console](https://console.hasura.io).

## Using the PostgreSQL connector

Check out the [Hasura docs here](https://hasura.io/docs/3.0/getting-started/build/connect-to-data/connect-a-source/?db=PostgreSQL) to get started with the PostgreSQL connector.

## Troubleshooting

Please [submit a Github issue](https://github.com/hasura/graphql-engine/issues/new)
if you encounter any problems!
