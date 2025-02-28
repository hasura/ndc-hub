# Quickstart Test Runner

## Overview

The Quickstart Test Runner is a tool designed to automate end-to-end testing for various connectors. It provides a standardized way to set up, run, and tear down test environments for different connectors.

## Structure

The test runner is organized as follows:

- `index.js`: The main entry point for running tests.
- `utils.js`: Contains utility functions used across the test suite.
- `fixtures/`: Directory containing test fixtures for each connector.

## Running Tests

Before running the tests, make sure you have the necessary dependencies installed:

```shell
bun install
```

The following Env Vars can be used to configure the test runner:

- `HASURA_DDN_PAT`: The PAT (Personal Access Token) for Hasura DDN (Can also use the service access tokens).
- `SELECTOR_PATTERN`: A glob pattern to select specific fixtures to test (Can be set to `postgres` to only test postgres connector test).
- `DDN_CLI_DIRECTORY`: The path to the directory containing the DDN CLI binary locally.
- `DDN_CLI_BINARY_NAME`: The name of the DDN CLI binary (Valid only if `DDN_CLI_DIRECTORY` is set, defaults to `cli-ddn-<os>-<arch><optional-.exe-suffix>` like `cli-ddn-linux-amd64` for linux/amd64 and `cli-ddn-windows-amd64.exe` for windows/amd64).
- `CLI_TAG`: The tag for the DDN CLI version to download (Required only if `DDN_CLI_DIRECTORY` is not set).
- `DDN_CLI_DOWNLOAD_URL`: The URL to download the DDN CLI binary (Required only if `DDN_CLI_DIRECTORY` is not set).

One of `DDN_CLI_DIRECTORY` or `CLI_TAG` or `DDN_CLI_DOWNLOAD_URL` must be set. `CLI_TAG` and `DDN_CLI_DOWNLOAD_URL` are mutually exclusive.

- `ENABLE_CLOUD_TESTS`: Set to `true` to enable cloud tests.

The following are only required if `ENABLE_CLOUD_TESTS` is set to `true`:

- `AUTH_ENDPOINT`: The authentication endpoint URL (For prod use `https://auth.pro.hasura.io`).
- `DATA_ENDPOINT`: The data endpoint URL (For prod use `https://data.pro.hasura.io`).
- `PROMPTQL_ENDPOINT`: The PromptQL endpoint URL (Required only if testing PromptQL, for prod use `https://data.promptql.pro.hasura.io/`).

To run the tests, use the following command:

```shell
bun run start
```

You can specify a particular directory or a glob pattern to test by setting the `SELECTOR_PATTERN` environment variable:

```shell
SELECTOR_PATTERN=postgres bun run start
```

## Adding a New Connector

To add a new connector to the test fixtures, follow these steps:

1. Create a new directory under `fixtures/` with the name of your connector.
2. Create `index.js` with the following structure:

```javascript
"use strict";

import {
  connector_init,
  connector_introspect,
  ddn,
  docker_compose_teardown,
  PROJECT_DIRECTORY,
  run_docker_start_detached,
  run_local_tests,
  supergraph_build_local,
  supergraph_init,
  track_all_commands,
  track_all_models,
  track_all_relationships,
} from "../../utils.js";

const NAME = "new_connector";

export async function setup(dir = PROJECT_DIRECTORY, ddnCmd = ddn()) {
  // Implement setup steps for your connector
}

export async function teardown(dir = PROJECT_DIRECTORY) {
  await docker_compose_teardown(dir);
}

export async function test_local(fixtureDir) {
  await run_local_tests(fixtureDir);
}

export async function test_cloud(
  dir = PROJECT_DIRECTORY,
  ddnCmd = ddn(),
  fixtureDir,
) {
  // Implement cloud tests if applicable
}
```

3. Implement the setup function with the necessary steps to initialize your connector, introspect the data source, track models, commands, and relationships, and build the supergraph.
4. Create a `snapshots/` directory inside your connector's fixture directory to store test snapshots along with the corresponding GraphQL requests and variables.

## Testing a New Connector

To test your newly added connector:

1. Ensure all dependencies are installed by running npm install in the `quickstart/test-runner` directory.
2. Run the tests for your specific connector by specifing the connector directory name using SELECTOR_PATTERN env variable:

```shell
SELECTOR_PATTERN=new_connector npm run start
```

3. Review the test output and fix any issues that arise.

## Continuous Integration

The test runner is integrated into the CI pipeline. The `.github/workflows/test-runner.yaml` file defines the CI job that runs these tests for different connectors and environments.
Update `.github/workflows/test-runner.yaml` file to include your new connector.

## Troubleshooting

If you encounter issues while adding or testing a new connector:

1. Ensure that your connector's setup function is correctly implemented and all necessary steps are included.
2. Verify that your connector's configuration is correct and all required environment variables are set.
