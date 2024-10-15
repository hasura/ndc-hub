# Connector registry Github packaging

> [!NOTE]
> This RFC is an update on the [Connector Package Distribution RFC](0002-distribution-gh.md).

## Introduction

This RFC proposes how a new connector version should be added to the `registry` folder to automatically be published. Publishing in this context means that the connector version will be available for public use in Hasura's DDN.

## File structure of the connectors `registry`

The packages field in the `metadata.json` file will be removed and replaced by a releases folder within the connector directory.

The releases folder will house a separate folder for each version of the connector, with each version folder containing a `connector-packaging.json` file.

This `connector-packaging.json` file will include all the necessary information to access the package definition.

The following directory structure for connector versions is proposed:

```
registry/<connector_namespace>/<connector_name>
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

The `registry` folder will contain a folder for each connector. Each connector folder will contain the following files:

- `logo.png`: The logo of the connector. The logo should be in PNG format.

- `metadata.json`: The metadata of the connector. Metadata format is described in the [Github Distribution RFC](./0002-distribution-gh.md).
- `README.md`: The README file of the connector. The README file should contain information about the connector, how to use it, and any other relevant information. The contents of the README file would be displayed in the landing page of the connector in the Hasura.
- `releases`: The releases folder will contain a folder for each version of the connector. Each version folder will contain a `connector-packaging.json` file. More details about the `connector-packaging.json` file are provided below.

NOTE: The `releases` folder should only be added for Hub connectors.
For example, `postgres-azure` connector should not have a `releases` folder as it is not a Hub connector.

### `connector-packaging.json`

Every connector version should have a package definition. The `connector-packaging.json`
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


## Publishing a new connector version

To publish a new connector version, follow these steps:

1. Create a new folder with the version number in the `releases` folder of the connector.
2. Create a `connector-packaging.json` file in the new folder with the relevant information.
3. Open a PR against the `main` branch of the repository.
4. You should see the `registry-update` workflow run on the PR. This workflow will validate the connector-packaging.json file and publish the new version to the registry if the validation is successful.
5. Once the workflow is successful, the new version of the connector will be available in the **Staging** Hasura DDN. Every new commit will overwrite the previous version of that connector in the staging DDN. So, feel free to push new commits to the PR to update the connector version in the staging DDN.
6. Once the PR is merged, the new version of the connector will be available in the **Production** Hasura DDN.

> [!NOTE]
> The `registry-update` workflow will only run on the PRs against the `main` branch of the repository.

> [!NOTE]
> Multiple connector versions can be published in the same PR. The `registry-update` workflow will publish all the versions in the PR to the registry.


## Updates to logo and README

If you want to update the logo or README of the connector, you can do so by opening a PR against the `main` branch of the repository.

The `registry-update` workflow will run on the PR and update the logo and README in the staging DDN.

Once the PR is merged, the logo and README will be updated in the production DDN.
