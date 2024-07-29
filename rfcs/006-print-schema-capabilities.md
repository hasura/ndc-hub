# RFC: `print-schema` command for all connectors

Currently, the only way to get the connector schema for CLI is to call `/schema` endpoint on a running NDC connector
instance. In case of local connector build, CLI builds and runs the respective NDC connector in a docker container and
calls the `/schema` endpoint on it. This is required by the CLI to build the DataConnectorLink which can then be used by
the engine to query the underlying data source. The new local CLI workflow is aimed at reducing the number of steps to
get a working API. In this workflow users should be able to update the DataConnectorLink without any dependency.

This RFC proposes to add a new optional command `print-schema` which prints out NDC schema to STDOUT of the underlying
data source. CLI can then call the plugin with relevant command to get the schema and construct DataConnectorLink from
the output. In case, the command is not implemented on plugin, CLI should rely on docker to get the schema.


## Changes to the packaging spec
```shell
packagingDefinition:
  type: ManagedDockerBuild/PrebuiltDockerImage
supportedEnvironmentVariables: []
commands:
  print-schema: hasura-ndc-plugin print-schema
```

```shell
ddn connector plugin --connector ./connector.yaml -- hasura-ndc-plugin print-schema

// Prints out schema to STDOUT
```