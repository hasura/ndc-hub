# Hasura Native Data Connector Hub: ndc-hub

This repository provides:

1. a registry of connectors and
2. resources to help build connectors to connect new and custom data sources to
   Hasura.

This allows Hasura users to instantly get a powerful Hasura GraphQL API
(pagination, filtering, sorting, relationships) with granular RLS style
authorization out of the box on any data-source (DBs, APIs).

> [!WARNING] NDC Hub (the set of connectors and the SDK to build new
> connectors) is currently in beta, and subject to large changes. It is
> shared here to provide an early preview of what can be expected for connector
> development & deployment in the future, and feedback is welcome! If you have
> any comments, please create an issue.

## Registry

There are some connectors you can start trying out today, and we'll gradually add more to this list:

1. [hasura/ndc-clickhouse](https://github.com/hasura/ndc-clickhouse)
1. [hasura/ndc-postgres](https://github.com/hasura/ndc-postgres)
2. [hasura/ndc-qdrant](https://github.com/hasura/ndc-qdrant)
3. [hasura/ndc-sendgrid](https://github.com/hasura/ndc-sendgrid)
4. [hasura/ndc-typescript-deno](https://github.com/hasura/ndc-typescript-deno)

## SDK & Guides

### Connector Developer Guide

The best way to get started developing Hasura native data connectors is to
[read the specification](http://hasura.github.io/ndc-spec/) and familiarise
yourself with the [reference
implementation](https://github.com/hasura/ndc-spec/tree/main/ndc-reference).

### Rust SDK

This repository provides a Rust crate to aid development of [Hasura Native Data
Connectors](https://hasura.github.io/ndc-spec/). Developers can implement a
trait, and derive an executable which can be used to run a connector which is
compatible with the specification.

In addition, this library adopts certain conventions which are not covered by
the current specification:

- Connector configuration
- State management
- Trace collection

#### Getting Started with the SDK

```sh
cd rust-connector-sdk
cargo build
```

#### Run the example connector

```sh
mkdir empty
cargo run --bin ndc_hub_example -- --configuration ./empty
```

Inspect the resulting (empty) schema:

```sh
curl http://localhost:8080/schema
```

(The default port, 8080, can be changed using `--port`.)

## Tracing

The serve command emits OTLP trace information. This can be used to see details
of requests across services.

To enable tracing you must:

- use the SDK option `--otlp-endpoint` e.g. `http://localhost:4317`,
- set the SDK environment variable `OTLP_ENDPOINT`, or
- set the `tracing` environment variable `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`.

For additional service information you can:

- Set `OTEL_SERVICE_NAME` e.g. `ndc_hub_example`
- Set `OTEL_RESOURCE_ATTRIBUTES` e.g. `key=value, k = v, a= x, a=z`

To view trace information during local development you can run a Jaeger server via Docker:

```
docker run --name jaeger -e COLLECTOR_OTLP_ENABLED=true -p 16686:16686 -p 4317:4317 -p 4318:4318 jaegertracing/all-in-one
```
