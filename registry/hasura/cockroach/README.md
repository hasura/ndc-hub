## Overview

`ndc-postgres` provides a Hasura Data Connector to the PostgreSQL database,
which can expose and run GraphQL queries via the Hasura v3 Project.

- [PostgreSQL Connector information in the Hasura Connectors directory](https://hasura.io/connectors/postgres)
- [GitHub repository](https://github.com/hasura/ndc-postgres)

The connector implements the [NDC Specification](https://hasura.github.io/ndc-spec/overview.html),
but does not currently support column relationship arguments in queries, or functions.

Visit the
[Hasura v3 Documentation](https://hasura.io/docs/3.0/native-data-connectors/postgresql) 
for more information.

The connector supports the [CockroachDB](https://www.cockroachlabs.com/) PostgreSQL-compatible database.

## Deployment

The connector is hosted by Hasura and can be used from the [Hasura v3 Console](https://console.hasura.io).

## Usage

Follow the [Quick Start Guide](https://hasura.io/docs/3.0/quickstart/) 
To use the PostgreSQL data connector from the [Hasura v3 Console](https://console.hasura.io).

## Troubleshooting

Please [submit a Github issue](https://github.com/hasura/graphql-engine/issues/new)
if you encounter any problems!
