# Hasura Native Data Connector Hub: ndc-hub

This repository provides:

1. a registry of connectors and
2. resources to help build connectors to connect new and custom data sources to
   Hasura.

This allows Hasura users to instantly get a powerful Hasura GraphQL API
(pagination, filtering, sorting, relationships) with granular RLS style
authorization out of the box on any data-source (DBs, APIs).

> **Warning:** NDC Hub (the set of connectors and the SDK to build new
> connectors) is currently in beta, and subject to large changes. It is shared
> here to provide an early preview of what can be expected for connector
> development & deployment in the future, and feedback is welcome! If you have
> any comments, please create an issue.

## Registry

The connectors currently supported all have an entry in the
[registry](/registry) folder.

## Guides

### SDKs

To get started quickly, we recommend using an SDK to build your own connector,
rather than starting from scratch.

- [Rust SDK]

[Rust SDK]: https://github.com/hasura/ndc-sdk-rs

### Connector Developer Guide

When developing Hasura Native Data Connectors, we recommend reading the [NDC
specification] and familiarizing yourself with the [reference
implementation][NDC reference].

[NDC specification]: http://hasura.github.io/ndc-spec/
[NDC reference]: https://github.com/hasura/ndc-spec/tree/main/ndc-reference
[Connector packaging]: ./connector-metadata-types/README.md
