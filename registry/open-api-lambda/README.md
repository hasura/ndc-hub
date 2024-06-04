# OpenAPI Lambda Connector

[![Docs](https://img.shields.io/badge/docs-v3.x-brightgreen.svg?style=flat)](https://hasura.io/docs/3.0/getting-started/overview/)
[![ndc-hub](https://img.shields.io/badge/ndc--hub-postgres-blue.svg?style=flat)](https://hasura.io/connectors)
[![License](https://img.shields.io/badge/license-Apache--2.0-purple.svg?style=flat)](LICENSE.txt)
[![Status](https://img.shields.io/badge/status-alpha-yellow.svg?style=flat)](./readme.md)

The OpenAPI Lambda Connector allows you to import APIs that have an OpenAPI/Swagger Documentation into the Hasura Supergraph. It works by creating the Types and API Calls required in Typescript and wrapping those API calls in functions. These functions can then be exposed as queries or mutations via the [NodeJS Lambda Connector](https://github.com/hasura/ndc-nodejs-lambda).

Functions that wrap GET requests are marked with `@readonly` annotation, and are exposed as GraphQL Queries by the [NodeJS Lambda Connector](https://github.com/hasura/ndc-nodejs-lambda). All other request types are exposed as GraphQL Mutations.

This Connector implements the [Data Connector Spec](https://github.com/hasura/ndc-spec)


- [Hasura V3 Documentation](https://hasura.io/docs/3.0)
- [NodeJS Lambda Connector](https://github.com/hasura/ndc-nodejs-lambda)

## Features

- Convert Open API/swagger documentation into Typescript functions compatible with NodeJS Lambda Connector
- Supported request types

| Request Type | Query | Path | Body | Headers |
| ------------ | ----- | ---- | ---- | ------- |
| GET          | y     | y    | NA   | y       |
| POST         | y     | y    | y    | y       |
| DELETE       | y     | y    | y    | y       |
| PUT          | y     | y    | y    | y       |
| PATCH        | y     | y    | y    | y       |

## Before you get Started

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Install the [CLI](https://hasura.io/docs/3.0/cli/installation/)
3. [Create a project](https://hasura.io/docs/3.0/getting-started/create-a-project)
4. Please ensure that you have Docker installed and the Docker Daemon is running.

## Usage

The Connector can be used with the DDN CLI. Use the `ddn dev` command to tell the CLI to automatically deploy, build and manage the Connector. Individual CLI Commands to build and deploy the Connector can be found [here](https://hasura.io/docs/3.0/cli/commands/)

## Useful Links

* [Open API Lambda GitHub Repository](https://github.com/hasura/ndc-open-api-lambda)
* [NodeJS Lambda GitHub Repository](https://github.com/hasura/ndc-nodejs-lambda)
* [README for Open API Lambda](https://github.com/hasura/ndc-open-api-lambda/blob/main/README.md)
* [Open API Lambda Changelog](https://github.com/hasura/ndc-open-api-lambda/blob/main/changelog.md)
