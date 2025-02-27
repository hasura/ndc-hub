"use strict";

import path from "path";
import fs from "fs";

import { minimatch } from "minimatch";
import {
  FIXTURES_DIRECTORY,
  PROJECT_DIRECTORY,
  runCommand,
  ddn,
  IS_CLOUD_TEST_ENABLED,
  setupDDNCLI,
  supergraph_init,
} from "./utils.js";

clear_project_dir();

await setupDDNCLI();

await login();
if (process.env.RUN_CLI_TESTS === "true") {
  let fixturesDir = process.env.FIXTURES_DIRECTORY || FIXTURES_DIRECTORY;
  await run_fixtures(fixturesDir);
} else {
  await run_ndc_tests();
}

async function login() {
  if (process.env.HASURA_DDN_PAT) {
    await runCommand(ddn(), [
      "auth",
      "login",
      "--pat",
      process.env.HASURA_DDN_PAT,
    ]);
  }
}

async function run_ndc_tests() {
  await runCommand(ddn(), ["ndc", "test"]);
}

async function run_fixtures(
  fixturesDir = path.resolve(CURRENT_DIRECTORY, "..", "..", "registry")
) {
  let selectorPattern = "*";
  if (process.env.SELECTOR_PATTERN) {
    selectorPattern = process.env.SELECTOR_PATTERN;
  }
  const namespaces = fs
    .readdirSync(fixturesDir, { withFileTypes: true })
    .filter((entry) => entry.isDirectory());
  for (const namespace of namespaces) {
    const namespaceDir = path.join(fixturesDir, namespace.name);
    console.log(`Testing namespace ${namespaceDir}`);
    const connectors = fs
      .readdirSync(namespaceDir, { withFileTypes: true })
      .filter((entry) => entry.isDirectory());

    for (const connector of connectors) {
      const connectorDir = path.join(namespaceDir, connector.name);
      console.log(`Testing connector ${connectorDir}`);
      const testConfigPath = path.join(connectorDir, "tests", "test-config.json");
      if (!fs.existsSync(testConfigPath))  {
        // TODO: add more data
        throw new Error(`No test-config.json found for ${connector.name}`);
      }
        const testConfig = JSON.parse(fs.readFileSync(testConfigPath, "utf8"));
        // TODO: validate this config
        // RUN the tests
        await supergraph_init(PROJECT_DIRECTORY, false, ddn());
        await connector_init(dir, ddnCmd, {
            connectorName: testConfig.connectorName,
            hubID: testConfig.hubID,
            port: testConfig.port || 8083,
            composeFile: "compose.yaml",
            envs: testConfig.envs || [],
          });
        if (testConfig.setupComposeFile) {
            await runCommand("docker", [
                "compose",
                "-f",
                path.join(path.dirname(testConfigPath), testConfig.setupComposeFile),
                "up",
                "--build",
                "-d",
                "--wait",
              ], {
                env: {
                  ...process.env,
                  CONNECTOR_CONTEXT_DIR: path.join(PROJECT_DIRECTORY, "app", "connector", connector.name),
                },
              });
        }
        // Small timeout to ensure all the pre-req is actually done
        await sleep(10000);
    }
  }
  const globalConfig = { projectName: "" };

  const directories = entries
    .filter((entry) => entry.isDirectory())
    .map((dir) => dir.name);

  const failedFixtures = [];
  const successfulFixtures = [];

  for (const dir of directories) {
    if (minimatch(dir, selectorPattern)) {
      clear_project_dir();
      let module;
      try {
        module = await import(
          pathToFileURL(path.join(fixturesDir, dir, "index.js"))
        );
      } catch (e) {
        console.error(`Error importing fixture ${dir}: ${e}`);
        continue;
      }
      try {
        console.log(`Setting up fixture "${dir}"`);
        await module["setup"](PROJECT_DIRECTORY, ddn(), globalConfig);

        console.log(`Testing fixture "${dir}" in local`);
        await module["test_local"](path.join(fixturesDir, dir), globalConfig);

        if (IS_CLOUD_TEST_ENABLED && module["test_cloud"]) {
          console.log(`Testing fixture ${dir} in cloud`);
          try {
            await module["test_cloud"](
              PROJECT_DIRECTORY,
              ddn(),
              path.join(fixturesDir, dir),
              globalConfig
            );
          } catch (err) {
            console.error(`Error testing fixture ${dir} in cloud: ${err}`);
            failedFixtures.push({
              name: dir,
              error: err,
              isCloud: true,
            });
          }
        }
        successfulFixtures.push(dir);
      } catch (e) {
        console.error(`Error testing fixture ${dir}: ${e}`);
        failedFixtures.push({
          name: dir,
          error: e,
        });
      } finally {
        try {
          await module["teardown"](PROJECT_DIRECTORY, globalConfig);
        } catch (err) {
          console.error(`Error tearing down fixture ${dir}: ${err}`);
        }
        clear_project_dir();
      }
    }
  }

  if (successfulFixtures.length > 0) {
    console.log("Successful fixtures: ", successfulFixtures);
  }

  if (failedFixtures.length > 0) {
    console.log("Failed fixtures: ", failedFixtures);
    throw new Error(`One or more tests failed`);
  }
}

function clear_project_dir(dir = PROJECT_DIRECTORY) {
  fs.rmSync(dir, { recursive: true, force: true });
  fs.mkdirSync(dir, { recursive: true });
}

function pathToFileURL(filepath) {
  let normalizedPath = filepath.replace(/\\/g, "/");
  if (process.platform === "win32") {
    normalizedPath = "/" + normalizedPath;
  }
  if (!normalizedPath.startsWith("/")) {
    normalizedPath = "/" + normalizedPath;
  }
  return `file://${normalizedPath}`;
}
