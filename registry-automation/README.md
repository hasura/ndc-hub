# Introduction

## Steps to runs

1. Consider the following `changed_files.json` file:
```json

{
    "added_files": [
        "registry/hasura/azure-cosmos/releases/v0.1.6/connector-packaging.json"
    ],
    "modified_files": [
        "registry/hasura/azure-cosmos/metadata.json"
    ],
    "deleted_files": []
}
```

2. You will require the following environment variables:

1. GCP_BUCKET_NAME
2. CLOUDINARY_URL
3. GCP_SERVICE_ACCOUNT_KEY
4. CONNECTOR_REGISTRY_GQL_URL
5. CONNECTOR_PUBLICATION_KEY
6. GCP_SERVICE_ACCOUNT_DETAILS



```bash


2. Run the following command from the `registry-automation` directory:


```bash
go run main.go ci --changed-files-path changed_files.json
```

## Steps to run the e2e helper

1. Run the following command from the `registry-automation` directory to run tests for changed files:

```bash
NDC_HUB_GIT_REPO_FILE_PATH=<path-to-repo-root> go run main.go e2e changed --changed-files-path changed_files.json
```
This command will fail if a newly added connector does not have a test config file.

2. Run the following command from the `registry-automation` directory to run tests for all connectors of all versions with tests setup:

```bash
NDC_HUB_GIT_REPO_FILE_PATH=<path-to-repo-root> go run main.go e2e all
```

3. Run the following command from the `registry-automation` directory to run tests for latest version of all connectors with tests setup:

```bash
NDC_HUB_GIT_REPO_FILE_PATH=<path-to-repo-root> go run main.go e2e latest
```

The output of these commands is a json configuration that can fed be into the [e2e-testing](./e2e-testing/) test runner. Pipe the output of the above commands into a file like `jobs.json` and use it as the input for the e2e-test runner.

4. To run the e2e-test runner (Install [bun](https://bun.sh/docs/installation) before running):
```bash
cd e2e-testing
bun install
NDC_HUB_GIT_REPO_FILE_PATH=<path-to-repo-root> TEST_JOB_FILE=<json-config-from-above> CLI_TAG=latest-staging  bun run start-ndc
```
