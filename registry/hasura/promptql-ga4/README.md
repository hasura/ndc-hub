# PromptQL GA4 Connector

The PromptQL GA4 Connector allows you to connect to Google Analytics and expose it to your DDN Project with PromptQL integration. This connector enables you to query GA4 data using natural language through PromptQL.

## Features

- Connect to Google Analytics properties
- Query GA4 data using natural language with PromptQL
- Support for dimensions, metrics, and date ranges
- Built-in authentication handling for Google Analytics API

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)
5. Google Analytics property with appropriate access permissions
6. Google Cloud Project with Google Analytics Data API enabled
7. Service account with access to your GA4 property

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes). You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the connector

### Initialize the connector

```bash
ddn connector init -i
```

Name the connector `promptql_ga4`.
Choose the hasura/promptql-ga4 connector from the list and enter the required environment variables.

### Required Environment Variables

- `GOOGLE_AUTH_CONFIG_FILEPATH`: Path to the Google Auth config file relative to the connector directory (optional, defaults to `google_auth_config.json`)

### Configure Authentication

1. Create a Google Cloud service account with Google Analytics Data API access
2. Grant the service account access to your GA4 property (Analytics Viewer role minimum)
3. Download the service account key file (JSON format)
4. Create a `google_auth_config.json` file in your connector directory with the following structure:

```json
{
  "property_id": "123456789",
  "domain": "example.com",
  "credentials": {
    "type": "service_account",
    "project_id": "your-project-id",
    "private_key_id": "your-private-key-id",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
    "client_email": "your-service-account@your-project.iam.gserviceaccount.com",
    "client_id": "your-client-id",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/your-service-account%40your-project.iam.gserviceaccount.com"
  }
}
```

Replace the values with your actual:
- `property_id`: Your GA4 property ID (numeric value)
- `domain`: Your website domain
- `credentials`: Your complete service account JSON credentials

### Introspect the connector

```bash
ddn connector introspect promptql_ga4
```

### Add the functions

```bash
ddn function add promptql_ga4 "*"
```

### Build and run locally

```bash
ddn supergraph build local
ddn run docker-start
```

Then visit the console and query your GA4 data through the GraphQL API.

## Supported Operations

- Query GA4 analytics data with dimensions and metrics
- Filter data by date range and other dimensions and metrics

## Example Queries

Query analytics data:
```graphql
query {
  analytics(
    date_range: {startDate: "2024-01-01", endDate: "2024-01-31"}
  ) {
    dimension_dimension_name
    metric_metric_name
  }
}
```
