# Stripe Connector

## Overview

The Stripe Data Connector provides an instant adapter for Engine v3 to request Stripe resources via GraphQL. This connector is built upon the [REST connector](https://github.com/hasura/ndc-rest) and [Stripe's OpenAPI Specification](https://github.com/stripe/openapi).

- [GitHub Repository](https://github.com/hasura/ndc-stripe)

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes).You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the Stripe connector

With the [context set](https://hasura.io/docs/3.0/cli/commands/ddn_context_set/) for an existing subgraph, initialize
the connector by choosing `hasura/stripe`. When the wizard runs, you'll also be prompted to enter the following env
vars necessary for your connector to function.

After the CLI initializes the connector, you'll need to:

- [Introspect](https://hasura.io/docs/3.0/cli/commands/ddn_connector_introspect) the source.
- Add your [models](https://hasura.io/docs/3.0/cli/commands/ddn_model_add) and
  [commands](https://hasura.io/docs/3.0/cli/commands/ddn_command_add).
- Create a [new build](https://hasura.io/docs/3.0/cli/commands/ddn_supergraph_build_local).
- Test it by [running your project along with the connector](https://hasura.io/docs/3.0/cli/commands/ddn_run#examples

See all available variables [here](https://github.com/hasura/ndc-stripe#environment-variables).

## License

The Hasura Stripe connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
