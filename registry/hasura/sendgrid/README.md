## Overview

This connector uses the SendGrid v3 API to:

* List email templates
* Send emails

## Prerequisites

1. [Create a SendGrid API account](https://signup.sendgrid.com/)
2. [Create a SendGrid API key](https://app.sendgrid.com/settings/api_keys)
2. Create a [Hasura Cloud account](https://console.hasura.io)
3. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
4. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
5. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes).You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the SendGrid Connector

Add the SendGrid connector to your DDN project by running

```
> ddn connector init -i
```

Select the SendGrid connector from the list and provide a name for the connector and your SendGrid API key.

Then you need to introspect the connector to get its schema:

```
> ddn connector introspect <connector name>
```

And then you can add all the SendGrid commands to your supergraph:

```
> ddn command add <connector name> "*"
```

You can now build your supergraph, run it locally, and open the Hasura Console to try it out:

```
> ddn supergraph build local
> ddn run docker-start
> ddn console --local
```
