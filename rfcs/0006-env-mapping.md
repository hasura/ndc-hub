# RFC: Connector environment variable mapping

To make the getting started [easier](https://docs.google.com/document/d/1ExL-rad3ck5lFz55AH7HCJwuucmyuSk0QoA8A_SJ-9g/edit#heading=h.5st3mzfx5bwk), the proposal is to make use of a single `.env` file (per environment, i.e. one per local, staging, prod, etc,.) that holds all the env variables of the entire system (all connectors and subgraphs). This can lead to potential collisions between connectors' Env vars (for e.g. I want to add two Postgres connectors and specify the `CONNECTION_URI` for each of them in the same `.env` file.). So, the idea is to add a prefix to the env vars in the `.env` file to avoid collisions. 

## Proposal
 - An optional `env.json` can be found at the at the root of connector context (the directory mounted to `/etc/connector`)
 - It is of the following format:
```typescript
type EnvJSON = {
    version: "v1"
    envMap?: EnvMap
}
type EnvMap = {
    [key: string]: Val
}
type Val = { value: string } | ValFromEnv
type ValFromEnv = {
    valueFromEnv: string
}
```
 - If present, the connector reads the env vars it requires (values currently specified in the [packaging spec](https://github.com/hasura/ndc-hub/blob/main/rfcs/0001-packaging.md) of each connector) and the ones mandated by [deployment spec](https://github.com/hasura/ndc-hub/blob/main/rfcs/0000-deployment.md) from the configuration specified in this file.
 - This ensures existing connectors are according to the spec. And new versions of connectors, must support using this `env.json` file.
 - DDN CLI generates this file during `ddn connector init` using the packaging spec.

## Sample `env.json` files

```json
{
    "version": "v1",
    "envMap": {
        "CONNECTION_URI": "postgres://user:pass@host:port/db",
        "HASURA_CONNECTOR_PORT": "8081",
    }
}
```

```json
{
    "version": "v1",
    "envMap": {
        "CONNECTION_URI": {
            "valueFromEnv": "APP_PG_CONNECTION_URI"
        },
        "HASURA_CONNECTOR_PORT": {
            "valueFromEnv": "APP_PG_HASURA_CONNECTOR_PORT"
        },
        "OTEL_EXPORTER_OTLP_ENDPOINT": {
            "valueFromEnv": "APP_PG_OTEL_EXPORTER_OTLP_ENDPOINT"
        }
    }
}
```






