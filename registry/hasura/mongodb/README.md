## Overview

`ndc-mongodb` provides a Hasura Data Connector to the MongoDB database,
which can expose and run GraphQL queries via the Hasura v3 Project.

- [GitHub repository](https://github.com/hasura/ndc-mongodb)

The connector implements the [NDC Specification](https://hasura.github.io/ndc-spec/overview.html),
but does not currently support mutations, column relationship arguments in queries, functions or procedures.

## Features

Below, you'll find a matrix of all supported features for the MongoDB data connector:

| Feature                                         | Supported | Notes |
| ----------------------------------------------- | --------- | ----- |
| Native Queries + Logical Models                 | ✅        |       |
| Simple Object Query                             | ✅        |       |
| Filter / Search                                 | ✅        |       |
| Filter by fields of Nested Objects              | ✅        |       |
| Filter by values in Nested Arrays               | ✅        |       |
| Simple Aggregation                              | ✅        |       |
| Aggregate fields of Nested Objects              | ❌        |       |
| Aggregate values of Nested Arrays               | ❌        |       |
| Sort                                            | ✅        |       |
| Sorty by fields of Nested Objects               | ❌        |       |
| Paginate                                        | ✅        |       |
| Collection Relationships                        | ✅        |       |
| Remote Relationships                            | ✅        |       |
| Relationships Keyed by Fields of Nested Objects | ❌        |       |
| Mutations                                       | ✅        | Provided by custom [Native Mutations](TODO) - predefined basic mutations are also planned |

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
