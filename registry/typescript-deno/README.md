## Overview

The Typescript (Deno) Connector allows a running connector to be inferred from a Typescript file (optionally with dependencies) and interpreted by [Deno](https://deno.com).

[github.com/hasura/ndc-typescript-deno](https://github.com/hasura/ndc-typescript-deno/tree/main#ndc-typescript-deno)

The connector runs in the following manner:

* The typescript sources are assembled
* Dependencies are fetched into a vendor directory
* Inference is performed and output to schema.json
* The functions are served via HTTP locally in the background
* The connector is started in the foreground responding to requests

It assumes that dependencies are specified in accordance with [Deno](https://deno.com) conventions.

## Typescript Functions Format

Your functions should be organised into a directory with `index.ts` file acting as the entrypoint.

```
// ./functions/index.ts

import { Hash, encode } from "https://deno.land/x/checksum@1.2.0/mod.ts";

/**
 * Returns an MD5 hash of the given password
 *
 * @param pw - Password string
 * @returns The MD5 hash of the password string
 * @pure This function should only query data without making modifications
 */
export function make_password_hash(pw: string): string {
    return new Hash("md5").digest(encode(pw)).hex();
}
```

* JSDoc comments and tags are exposed in the schema
* Async, and normal functions are both supported
* Only exported functions are exposed
* Functions tagged with `@pure` annotations are exposed as functions
* Those without `@pure` annotations are exposed as procedures

## Deployment

You will need:

* [V3 CLI](https://github.com/hasura/v3-cli) (With Logged in Session)
* [Connector Plugin](https://hasura.io/docs/latest/hasura-cli/connector-plugin/)
* Secret service token
* A configuration file

Your functions directory should be added as a volume to `/functions`

```
--volume ./my-functions:/functions
```

Create the connector:

```
hasura3 connector create my-cool-connector:v1 \
  --github-repo-url https://github.com/hasura/ndc-typescript-deno/tree/v0.8 \
  --config-file config.json \
  --volume ./my-functions:/functions \
  --env SERVICE_TOKEN_SECRET=MY-SERVICE-TOKEN
```

Monitor the deployment status by name:

```
hasura connector status my-cool-connector:v1
```

List your connector with its deployed URL:

```
hasura connector list
my-cool-connector:v1 https://connector-9XXX7-hyc5v23h6a-ue.a.run.app active
```

See [the Typescript Deno SendGrid repository](https://github.com/hasura/ndc-sendgrid-deno)
for an example of what a project structure that uses a connector could look like.

## Usage

Include the connector URL in your Hasura V3 project metadata (hml format).
Hasura cloud projects must also set a matching bearer token:

```yaml
kind: DataSource
name: sendgrid
dataConnectorUrl:
  url: 'https://connector-9XXX7-hyc5v23h6a-ue.a.run.app'
auth:
  type: Bearer
  token: "SUPER_SECRET_TOKEN_XXX123"
```

While you can specify the token inline as above, it is recommended to use the Hasura secrets functionality for this purpose:

```yaml
kind: DataSource
name: sendgrid
dataConnectorUrl:
  url: 'https://connector-9XXX7-hyc5v23h6a-ue.a.run.app'
auth:
  type: Bearer
  token:
    valueFromSecret: CONNECTOR_TOKEN
```


## Troubleshooting

Please [submit a Github issue](https://github.com/hasura/ndc-typescript-deno/issues/new)
if you encounter any problems!
