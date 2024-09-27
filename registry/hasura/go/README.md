## Overview

The Go connector allows you to expose Go functions as NDC functions/procedures for use in your Hasura DDN subgraphs. The
connector provides a boilerplate with NDC Go SDK and a generation tool to generate NDC schema and DRY functions from Go
code.

- [GitHub Repository](https://github.com/hasura/ndc-sdk-go)

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes).You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the Go connector

Check out the [Hasura docs here](https://hasura.io/docs/3.0/business-logic/go#add-the-go-connector-to-a-project) to get
started with the Go connector.

## Compatibility

| Go Version | SDK Version |
| ---------- | ----------- |
| 1.21+      | v1.x        |
| 1.19+      | v0.x        |

## More Information

- [Hasura DDN Documentation](https://hasura.io/docs/3.0/business-logic/go)
- [GitHub Repository](https://github.com/hasura/ndc-sdk-go/tree/main/cmd/hasura-ndc-go)
