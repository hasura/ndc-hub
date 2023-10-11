## Overview

The Typescript (Deno) Connector allows a running connector to be inferred from a Typescript file (optionally with dependencies) and interpreted by [Deno](https://deno.com).

[github.com/hasura/ndc-typescript-deno](https://github.com/hasura/ndc-typescript-deno/tree/main#ndc-typescript-deno)

The connector runs in the following manner:

* Dependencies are fetched
* Inference is performed
* The functions are served via the [connector protocol](https://github.com/hasura/ndc-spec/tree/main#ndc-specification)

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
* Optional parameters are supported
* Exceptions can be thrown and will be reported to the user

## Function Development

For the best user-experience you should develop your functions in the following manner:

* Have [Deno](https://deno.com) installed
* Have [VSCode](https://code.visualstudio.com) installed
* Have the [Deno VSCode extension](https://marketplace.visualstudio.com/items?itemName=denoland.vscode-deno) installed

An example session:

```
> tree
.
├── config.json
├── functions
    ├── index.ts
    └── schema.json

> cat config.json 
{
  "functions": "./functions/index.ts",
  "vendor": "./vendor",
  "preVendor": true,
  "schemaMode": "INFER",
  "schemaLocation": "./functions/schema.json"
}

> cat functions/index.ts 

export function hello(): string {
  return "hello world";
}

function foo() {
}

> deno run --allow-sys --allow-run --allow-net --allow-read --allow-write --allow-env --watch --check https://deno.land/x/hasura_typescript_connector@0.9/mod.ts serve --configuration ./config.json
Watcher Process started.
Check https://deno.land/x/hasura_typescript_connector@0.9/mod.ts
Running Connector.start
Check file:///Users/me/hasura/ndc-typescript-deno/scratch/deno_land_test/functions/index.ts
Inferring schema with map location ./vendor
Vendoring dependencies: /Users/me/bin/binaries/deno vendor --output /Users/me/hasura/ndc-typescript-deno/scratch/deno_land_test/vendor --force /Users/me/hasura/ndc-typescript-deno/scratch/deno_land_test/functions/index.ts
Skipping non-exported function: foo
Writing schema to ./functions/schema.json
{"level":30,"time":1697018006809,"pid":89762,"hostname":"spaceship.local","msg":"Server listening at http://0.0.0.0:8100"}
```

Once your connector is running locally you can use the `hasura3 tunnel` commands to make it available to your cloud projects for testing.

If you are happy with its behaviour you can deploy your connector vis `hasura3 connector` commands.


## Deployment

You will need:

* [V3 CLI](https://github.com/hasura/v3-cli) (with a logged in session)
* [Connector Plugin](https://hasura.io/docs/latest/hasura-cli/connector-plugin/)
* A connector configuration file
* Secret service token (optional)

Your functions directory should be added as a volume to `/functions`

```
--volume ./my-functions:/functions
```

Create the connector:

```
hasura3 connector create my-cool-connector:v1 \
  --github-repo-url https://github.com/hasura/ndc-typescript-deno/tree/v0.9 \
  --config-file config.json \
  --volume ./functions:/functions \
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
kind: DataConnector
version: v1
definition:
  name: sendgrid
  url:
    singleUrl: 'https://connector-9XXX7-hyc5v23h6a-ue.a.run.app'
  headers:
    Authorization:
      value: "Bearer SUPER_SECRET_TOKEN_XXX123"
```

While you can specify the token inline as above, it is recommended to use the Hasura secrets functionality for this purpose:

```yaml
kind: DataConnector
version: v1
definition:
  name: sendgrid
  url:
    singleUrl: 'https://connector-9XXX7-hyc5v23h6a-ue.a.run.app'
  headers:
    Authorization:
      valueFromSecret: BEARER_TOKEN_SECRET
```

(NOTE: This will require that the secret includes the `Bearer ` prefix.)


## Troubleshooting

Please [submit a Github issue](https://github.com/hasura/ndc-typescript-deno/issues/new)
if you encounter any problems!
