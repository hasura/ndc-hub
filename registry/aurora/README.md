# Hasura PostgreSQL (Aurora) Connector

<a href="https://hasura.io/"><img src="https://github.com/hasura/ndc-postgres/blob/main/docs/logo.png" align="right" width="200"></a>

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/connectors/postgresql)
[![Latest release](https://img.shields.io/github/v/release/hasura/ndc-postgres)](https://github.com/hasura/ndc-postgres/releases/latest)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-postgres-blue.svg?style=flat)](https://hasura.io/connectors/postgres)

The Hasura PostgreSQL Connector allows for connecting to a PostgreSQL database giving you an instant
GraphQL API on top of your PostgreSQL data.

As much as possible we attempt to provide explicit support for database projects that identify as being derived from PostgreSQL such as [AWS RDS and Aurora PostgreSQL](https://aws.amazon.com/rds/aurora/).

This connector is built using the [Rust Data Connector SDK](https://github.com/hasura/ndc-sdk-rs)
and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

- [Connector information in the Hasura Hub](https://hasura.io/connectors/postgres)
- [Hasura V3 Documentation](https://hasura.io/docs/3.0)

## Features

Below, you'll find a matrix of all supported features for the PostgreSQL connector:

| Feature                | Supported |
| ---------------------- | --------- |
| Native Queries         | ✅        |
| Native Mutations       | ✅        |
| Simple Object Query    | ✅        |
| Filter / Search        | ✅        |
| Simple Aggregation     | ✅        |
| Sort                   | ✅        |
| Paginate               | ✅        |
| Table Relationships    | ✅        |
| Views                  | ✅        |
| Mutations              | ✅        |
| Distinct               | ✅        |
| Enums                  | ✅        |
| Default Values         | ✅        |
| User-defined Functions | ❌        |

## Using the PostgreSQL connector

Hasura DDN's [Getting Started](https://hasura.io/docs/3.0/getting-started/build/connect-to-data/connect-a-source?db=PostgreSQL)
guide contains information about how to use the PostgreSQL connector as part of a Hasura DDN project.

## Support & Troubleshooting

The documentation and community will help you troubleshoot most issues.
If you have encountered a bug or need to get in touch with us, you can contact us using one of the following channels:

- Support & feedback: [Discord](https://discord.gg/hasura)
- Issue & bug tracking: [GitHub issues](https://github.com/hasura/graphql-engine/issues)
- Follow product updates: [@HasuraHQ](https://twitter.com/hasurahq)
- Talk to us on our [website chat](https://hasura.io)

We are committed to fostering an open and welcoming environment in the community.
Please see the [Code of Conduct](https://github.com/hasura/ndc-postgres/blob/main/docs/code-of-conduct.md).
If you want to report a security issue, please [read this](https://github.com/hasura/ndc-postgres/blob/main/docs/security.md).

## Documentation

View the full documentation for the connector [here](https://github.com/hasura/ndc-postgres/blob/main/docs/readme.md).

### Production

See the [production guide](https://github.com/hasura/ndc-postgres/blob/main/docs/production.md) for details about production setup.

### Development

See the [development guide](https://github.com/hasura/ndc-postgres/blob/main/docs/development.md) for details about development workflows, tooling, and code structure.

## Contributing

`ndc-postgres` is still in early stages of development and we are currently not accepting contributions.

## License

The Hasura PostgreSQL Connector is available under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0) (Apache-2.0).
