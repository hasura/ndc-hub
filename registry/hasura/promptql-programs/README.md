# NDC PromptQL Programs Connector
PromptQL Programs Connector allows you to invoke PromptQL Programs (Automations) as commands. 

## Current Limitations
- If the Input or output schema of a program results in types that are [not supported](https://github.com/hasura/ndc-nodejs-lambda?tab=readme-ov-file#unsupported-types) by the NDC Nodejs lambda connector, the program will not be available to track as a command. (For eg. If one of the field in the program input is an enum, it results in an Union type for that field in TS. As Union types are not supported by the NDC Nodejs lambda connector (without the the relaxed types hack, which anyway is not usable with PromptQL), the program will not be available to track as a command. To get around this, change the schema to use a string instead of enum and add the possible values of this enum as a description in the Command/Object Type for PromptQL to use.)

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes).You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).


## Using the connector
- Create a PromptQL Project
In a fresh directory, do the following:
```bash
ddn supergraph init <my-project> --with-promptql && cd <my-project>
``` 
- Init the connector
```bash
ddn connector init -i
``` 
Name the connector `promptql_programs`.
Choose the hasura/promptql-programs connector from the list and enter the required environment variables.
Docs for getting execute program API endpoint details - https://promptql.io/docs/promptql-apis/execute-program-api/
- Add Argument presets to the connector for header forwarding

Add the following to `promptql_programs.hml` file under `definition` section to enable header forwarding. (Even if you don't need header forwarding add a dummy configuration)
```yaml
argumentPresets:
    - argument: headers
      value:
        httpHeaders:
          forward:
            - "Authorization" # Modify as per what all headers you need to forward. If you do not need header forwarding, still put a dummy header value here.
          additional: {}
```
- Add the Automation JSON Artifact

Download the Automation as a JSON artifact from the PromptQL console and place it in the `app/connector/promptql_programs/programs` directory.
- Introspect the connector
```bash
ddn connector introspect promptql_programs
```
- Add the commands
```bash
ddn command add promptql_programs "*"
```

- Do a local build
```bash
ddn supergraph build local
```

- Run PromptQL locally
```bash
ddn run docker-start
```
Then visit the console and invoke the sum automation via Playground
