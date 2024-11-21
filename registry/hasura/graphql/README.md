## Overview

The Hasura GraphQL Connector allows for connecting to a GraphQL API and bringing it into Hasura DDN supergraph as a
single unified API. It can also be used to bring in your current Hasura v2 graphQL API into Hasura DDN and our
recommended approach is to create a new subgraph for the v2 API.

For Hasura v2 users, this functionality is the replacement of
[remote schemas](https://hasura.io/docs/latest/remote-schemas/overview/) functionality in v3 (DDN).

The `ndc-graphql` data connector is open source and can be found in the
[ndc-graphql GitHub repository](https://github.com/hasura/ndc-graphql).

## Features

Below, you'll find a matrix of all supported features for the GraphQL connector:

| Feature                | Supported | Notes |
| ---------------------- | --------- | ----- |
| Queries                | ✅        | All features that v3 engine currently supports
| Mutations              | ✅        |
| Header Passthrough     | ✅        | Entire headers can be forwarded
| Subscriptions          | ❌        |
| Unions                 | ❌        | Can be brought in via scalar types
| Interfaces             | ❌        |
| Relay API              | ❌        |
| Directives             | ❌        | @cached, Apollo directives

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes).You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the GraphQL connector

Check out the
[Hasura docs here](https://hasura.io/docs/3.0/getting-started/build/connect-to-data/connect-a-source/?db=GraphQL) to get
started with the GraphQL connector.

## Deployment

The connector is hosted by Hasura and can be used from the [Hasura v3 CLI](https://hasura.io/docs/3.0/cli/overview/) and
[Console](https://console.hasura.io). Please follow either the
[Quick Start Guide](https://hasura.io/docs/3.0/getting-started/overview/) or
[deploying the connector](https://hasura.io/docs/3.0/connectors/deployment).

## Troubleshooting

Please [submit a Github issue](https://github.com/hasura/graphql-engine/issues/new)if you encounter any problems!
