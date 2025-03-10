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
[Connector metadata packaging]: ./connector-metadata-types/README.md

### Adding e2e tests for your connector

All new connector releases to NDC hub MUST have e2e tests. The e2e tests are
run in the CI pipeline for every connector release. The e2e tests are run using
the [e2e-testing](./registry-automation/e2e-testing/) test runner. 

[This](https://github.com/hasura/ndc-hub/pull/485/files) PR to add a new version of python connector with e2e tests can be used as an example of how to add e2e tests for new connector releases.

You need to add the following configuration to enable e2e tests:

- Add a `tests/test-config.json` file with the config for your connector. See [test-config.json](./registry/hasura/mysql/tests/test-config.json) for an example.
  - The `hub_id` is the value hub identifier for your connector. It is of the format `<namespace>/<connector-name>` (for e.g. `hasura/mysql`).
  - Add an optional `port` in the `test-config.json` file if your connector runs on a non-default port. The default port is `8083`.
  - Add an optional `envs` in the `test-config.json` file if your connector requires any environment variables during init. See [test-config.json](./registry/hasura/mysql/tests/test-config.json) for an example.
  - Add snapshots for your connector in the `tests/snapshots` folder. Each snapshot is a folder (for e.g. `query1`) with a `request.graphql` and a `response.json` file. The request is the GraphQL query to be run and the response is the expected response. You can add an optional `variables.json` file if the query has variables. See [snapshots](./registry/hasura/mysql/tests/snapshots) for an example.
  - Add an optional `setup_compose_file_path` in the `test-config.json` file if you need to setup any additional services for your connector. See [compose.yaml](./registry/hasura/mysql/tests/compose.yaml) for an example. This compose file is invoked before doing connector introspect. This can hence be used to setup the connector like starting a local database for the connector to connect to (like [here](./registry/hasura/mysql/tests/compose.yaml)) or setup your connector config(like [here](./registry/hasura/nodejs/tests/compose.yaml)).
    - Note that every service you setup must have a [healthcheck](https://docs.docker.com/reference/compose-file/services/#healthcheck) defined. The `healthcheck` is used to wait for the service to be ready before running the tests (like [here](./registry/hasura/mysql/tests/compose.yaml)).
    - The Environment variable `CONNECTOR_CONTEXT_DIR` is available in the compose file. This is the path to the connector directory in the local project during the test. You can mount this directory to your service to access the connector config and make any changes/setup your connector before introspection (like [here](./registry/hasura/nodejs/tests/compose.yaml)).
    - If the service in the compose.yaml is a shortlived service (like [here](./registry/hasura/nodejs/tests/compose.yaml)  which is used to modify the `functions.ts`  and exits), you can add a `restart: "no"` to the service to prevent it from restarting after it exits and the recommendation is to add a small `sleep` at the end of the service to ensure that the service exits after the changes are made.
- Refer this this `test-config.json` file in the `connector-packaging.json` file of your connector release. If you followed the above file structure, it will be
```diff
{
    "version": "v0.0.1",
    "uri": "your-connector-package-url",
    "checksum": {
        "type": "sha256",
        "value": "2cd3584557be7e2870f3488a30cac6219924b3f7accd9f5f473285323843a0f4"
    },
    "source": {
        "hash": "c32adbde478147518f65ff465c40a0703239288a"
+    },
+    "test": {
+        "test_config_path": "../../tests/test-config.json"
+    }
}
```
