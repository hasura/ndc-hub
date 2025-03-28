# Quickstart Test Runner

## Overview

The Quickstart Test Runner is a tool designed to automate end-to-end testing for various connectors. It provides a standardized way to set up, run, and tear down test environments for different connectors.

## Running NDC HUB tests

Before running the tests, make sure you have the necessary dependencies installed:

1. Install [bun](https://bun.sh/docs/installation)
2. Then install project dependencies

```shell
bun install
```

The following Env Vars can be used to configure the test runner:

- `HASURA_DDN_PAT`: (REQUIRED)The PAT (Personal Access Token) for Hasura DDN (Can also use the service access tokens).
- `DDN_CLI_DIRECTORY`: The path to the directory containing the DDN CLI binary locally.
- `DDN_CLI_BINARY_NAME`: The name of the DDN CLI binary (Valid only if `DDN_CLI_DIRECTORY` is set, defaults to `cli-ddn-<os>-<arch><optional-.exe-suffix>` like `cli-ddn-linux-amd64` for linux/amd64 and `cli-ddn-windows-amd64.exe` for windows/amd64).
- `CLI_TAG`: The tag for the DDN CLI version to download (Required only if `DDN_CLI_DIRECTORY` is not set).
- `DDN_CLI_DOWNLOAD_URL`: The URL to download the DDN CLI binary (Required only if `DDN_CLI_DIRECTORY` is not set).
- `NDC_HUB_GIT_REPO_FILE_PATH`: (REQUIRED)The path to the root of the ndc-hub repo.
- `TEST_JOB_FILE`: (REQUIRED)The path to the JSON file containing the test job configuration (This can be generated by the e2e [helper](../README.md)).The file must be a json of type `TestJob[]` defined in [ndc.ts](./ndc.ts).

One of `DDN_CLI_DIRECTORY` or `CLI_TAG` or `DDN_CLI_DOWNLOAD_URL` must be set. `CLI_TAG` and `DDN_CLI_DOWNLOAD_URL` are mutually exclusive.

The following are only required if the job is configured to run cloud tests:

- `AUTH_ENDPOINT`: The authentication endpoint URL (For prod use `https://auth.pro.hasura.io`).
- `DATA_ENDPOINT`: The data endpoint URL (For prod use `https://data.pro.hasura.io`).
- `PROMPTQL_ENDPOINT`: The PromptQL endpoint URL (Required only if testing PromptQL, for prod use `https://data.promptql.pro.hasura.io/`).

To run the tests, use the following command with right env vars set:

```shell
bun run start-ndc
```
