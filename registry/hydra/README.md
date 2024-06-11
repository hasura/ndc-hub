# Ory Hydra Connector

## Overview

Ory Hydra connector provides an instant adapter for Engine v3 to request Hydra API resources via GraphQL. This connector is built upon the [REST connector](https://github.com/hasura/ndc-rest) and [Ory Hydra's REST API Specification](https://raw.githubusercontent.com/ory/hydra/master/internal/httpclient/api/openapi.yaml).

## Usage

Set the following environment variables and start the container. See all available variables [here](https://github.com/hasura/ndc-hydra/tree/main?tab=readme-ov-file#environment-variables).

| Name                    | Description             | Default Value         |
| ----------------------- | ----------------------- | --------------------- |
| HYDRA_PUBLIC_SERVER_URL | Public Hydra server URL | http://localhost:4444 |
| HYDRA_ADMIN_SERVER_URL  | Admin Hydra server URL  | http://localhost:4445 |

## Supported versions

| Connector Version | Hydra version |
| ----------------- | ------------- |
| `v0.x`            | `v1.x`        |
| `v1.x`            | `v2.x`        |
