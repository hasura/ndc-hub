## Overview

The Hasura GraphQL Connector allows for connecting to a GraphQL API and bringing it into Hasura DDN supergraph as a
single unified API. It can also be used to bring in your current Hasura v2 GraphQL API into Hasura DDN.

For Hasura v2 users, this functionality is the replacement of
[Remote Schemas](https://hasura.io/docs/latest/remote-schemas/overview/) functionality in DDN.

## Features

Below, you'll find a matrix of all supported features for the GraphQL connector:

| Feature            | Supported | Notes                                          |
| ------------------ | --------- | ---------------------------------------------- |
| Queries            | ✅        | All features that v3 engine currently supports |
| Mutations          | ✅        |
| Header Passthrough | ✅        | Entire headers can be forwarded                |
| Subscriptions      | ❌        |
| Unions             | ❌        | Can be brought in via scalar types             |
| Interfaces         | ❌        |
| Relay API          | ❌        |
| Directives         | ❌        | @cached, Apollo directives                     |

## Build on Hasura DDN

[Get started](https://hasura.io/docs/3.0/how-to-build-with-ddn/with-graphql) by connecting your GraphQL endpoint to a Hasura DDN project.

## Fork the connector

You can fork the [connector's repo](https://github.com/hasura/ndc-graphql) and iterate on it yourself.
