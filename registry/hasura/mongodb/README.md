## Overview

`ndc-mongodb` provides a Hasura Data Connector to the MongoDB database,
which can expose and run GraphQL queries via the Hasura v3 Project.

- [GitHub repository](https://github.com/hasura/ndc-mongodb)

The connector implements the [NDC Specification](https://hasura.github.io/ndc-spec/overview.html),
but does not currently support mutations, column relationship arguments in queries, functions or procedures.

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

## Using the MongoDB connector

Check out the [Hasura docs here](https://hasura.io/docs/3.0/getting-started/build/connect-to-data/connect-a-source/?db=MongoDB) to get started with the MongoDB connector.

## Troubleshooting

Please [submit a Github issue](https://github.com/hasura/graphql-engine/issues/new)
if you encounter any problems!
