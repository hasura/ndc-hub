## Overview
The Hasura MongoDB Connector allows for connecting Hasura to a MongoDB database giving you an instant GraphQL API on top of your MongoDB data.

Data Connectors are the way to connect the Hasura Data Delivery Network (DDN) to external data sources. A data connector is an HTTP service that exposes a set of APIs that Hasura uses to communicate with the data source. Data connectors are built to conform to the [NDC Specification](https://hasura.github.io/ndc-spec/overview.html) using one of Hasura's available SDKs. The data connector is responsible for interpreting work to be done on behalf of the Hasura Engine, using the native query language of the data source.

MongoDB is a popular, open-source NoSQL database system designed to store and manage vast amounts of data. It uses a document-oriented data model, where data is stored in flexible, JSON-like documents. MongoDB is known for its scalability, high performance, and ease of use, making it well-suited for a wide range of applications. It's often used in web development, mobile apps, real-time analytics, and other modern data-driven applications.

`ndc-mongodb` provides a Hasura Data Connector to the MongoDB database,
which can expose and run GraphQL queries via the Hasura v3 Project.

- [GitHub repository](https://github.com/hasura/ndc-mongodb)

## Deployment

The connector is hosted by Hasura and can be used from the [Hasura v3 Console](https://console.hasura.io).

## Usage

Follow the [Quick Start Guide](https://hasura.io/docs/3.0/quickstart/) 
to use the MongoDB data connector from the [Hasura v3 Console](https://console.hasura.io).

To add the `ndc-mongodb` connector to a Hasura DDN project run the following command:
```
ddn add connector-manifest <connector name> --subgraph app --hub-connector hasura/mongodb --type cloud
```

## Troubleshooting

Please [submit a Github issue](https://github.com/hasura/graphql-engine/issues/new)
if you encounter any problems!
