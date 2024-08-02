# Connector registry Github packaging

This is a Work-In-Progress document. Please provide any feedback you wish to contribute via Github comments and suggestions.

## Introduction

This RFC proposes how a new connector version should be added to the `registry` folder to automatically be published. Publishing here
means that the connector version will be available for use in Hasura's DDN.

## Directory structure

The following directory structure for connector versions is proposed:

```
registry/<connector_name>
├── logo.png
├── metadata.json
├── README.md
└── releases
    ├── v0.0.1
    │   └── connector-packaging.json
    ├── v0.0.2
    │   └── connector-packaging.json
    ├── v0.0.3
    │   └── connector-packaging.json
    ├── v0.0.4
    │   └── connector-packaging.json
    ├── v0.0.5
    │   └── connector-packaging.json
    ├── v0.0.6
    │   └── connector-packaging.json
    ├── v0.1.0
    │   └── connector-packaging.json
    └── v1.0.0
        └── connector-packaging.json
```

### `connector-packaging.json`

Every connector version has a package defintion, as specified here(TODO: Link to the packaging RFC). The `connector-packaging.json`
file should contain the relevant information to access the package definition.

```json
{
    "version": "0.0.1",
    "uri": "https://github.com/hasura/ndc-mongodb/releases/download/v0.0.1/connector-definition.tgz",
    "checksum": {
        "type": "sha256",
        "value": "2cd3584557be7e2870f3488a30cac6219924b3f7accd9f5f473285323843a0f4"
    },
    "source": {
        "hash": "c32adbde478147518f65ff465c40a0703239288a"
    }
}
```

The following fields are required:

- `version`: The version of the connector.
- `uri`: The URI to download the connector package. The package should be a tarball containing the connector package definition and the URL should be accessible without any authentication.
- `checksum`: The checksum of the connector package. The checksum should be calculated using the `sha256` algorithm.
- `source`: The source of the connector package. The `hash` field should contain the commit hash of the source code that was used to build the connector package.
