# NDC Spec Version in Packaged Connector Metadata

## Problem Statement

The Hasura 3.0 CLI needs to differentiate between connectors implementing different NDC spec versions (0.1, 0.2, and future versions). Currently, the spec version is only available in the connector's schema response, but the CLI cannot access this information without running the connector, creating a dependency cycle.

## Proposed Solution

Include the NDC spec version directly within the `./hasura-connector/connector-metadata.yaml` file, allowing the CLI to identify the connector's compatibility without execution.

## Implementation
Add two new top-level fields to the connector metadata YAML:

> Note: If `version` is `v1` or undefined, the `ndcSpecVersion` is assumed to be `0.1`.

```diff
++version: v2
++ndcSpecVersion: 0.2
packagingDefinition:
  type: PrebuiltDockerImage
  dockerImage: "ghcr.io/hasura/ndc-athena-jdbc:v1.0.0"
...
```

## Benefits
- Enables the CLI to determine compatibility without executing the connector
- Creates a clear separation of concerns between connector configuration and NDC specification version
- Supports graceful handling of version-specific features and behaviors
- Provides forward compatibility for future spec versions
