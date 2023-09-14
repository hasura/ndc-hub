## Overview

The Typescript (Deno) Connector allows a running connector to be inferred from a Typescript file (optionally with dependencies) and interpreted by [Deno](https://deno.com).

https://github.com/hasura/ndc-typescript-deno/tree/main#ndc-typescript-deno

The connector runs in the following manner:

* The typescript sources are assembled
* Dependencies are fetched into a vendor directory
* Inference is performed and output to schema.json
* The functions are served via HTTP locally in the background
* The connector is started in the foreground responding to requests

It assumes that dependencies are specified in accordance with [Deno](https://deno.com) conventions.

## Typescript Functions Format

Your functions should be organised into a directory with one file acting as the entrypoint.

An example could be as follows - `functions/main.ts`:

```
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

## Deployment

You will need:

* [V3 CLI](https://github.com/hasura/v3-cli) (With Logged in Session)
* [Connector Plugin](https://hasura.io/docs/latest/hasura-cli/connector-plugin/)
* Secret service token
* A configuration file

The configuration file format needs at a minimum
a `typescript_source` referenced which matches the main
typescript file as mounted with the `--volume` flag.

```
{"typescript_source": "/functions/main.ts"}
```

Create the connector:

> hasura3 connector create my-cool-connector:v1 \\
> --github-repo-url https://github.com/hasura/ndc-typescript-deno/tree/main \\
> --config-file config.json \\
> --volume ./functions:/functions \\
> --env SERVICE_TOKEN_SECRET=MY-SERVICE-TOKEN

Monitor the deployment status by name:

> hasura connector status my-cool-connector:v1

List your connector with its deployed URL:

> hasura connector list

```
my-cool-connector:v1 https://connector-9XXX7-hyc5v23h6a-ue.a.run.app active
```

## Usage

Include the connector URL in your Hasura V3 project metadata:

```json
[
  {
      "kind": "dataSource",
      "name": "md5hasher",
      "dataConnectorUrl": "https://connector-9XXX7-hyc5v23h6a-ue.a.run.app",
      "schema": {}
  }
  ...
]
```

## Troubleshooting

Please [https://github.com/hasura/ndc-typescript-deno/issues/new](submit a Github issue)
if you encounter any problems!
