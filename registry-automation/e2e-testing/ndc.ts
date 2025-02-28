"use strict";

import path from "path";
import fs from "fs";
import { minimatch } from "minimatch";
import {
  PROJECT_DIRECTORY,
  runCommand,
  ddn,
  setupDDNCLI,
  supergraph_init,
  CURRENT_DIRECTORY,
  connector_init,
  connector_introspect,
  track_all_models,
  track_all_commands,
  track_all_relationships,
  supergraph_build_local,
  run_docker_start_detached,
  docker_compose_teardown,
  run_local_tests,
  sleep,
  type FailedFixture,
  type TestConfig,
  clear_project_dir,
  login,
  project_init,
  type GlobalConfig,
  printPAT,
  get_project_id,
  supergraph_build_create,
  run_cloud_tests,
  project_delete,
  validateTestConfig,
} from "./utils";

clear_project_dir();

await setupDDNCLI();

await login();
await run_fixtures();

async function run_fixtures(
  fixturesDir: string = path.resolve(CURRENT_DIRECTORY, "..", "..", "registry"),
): Promise<void> {
  let selectorPattern: string = "*";
  if (process.env.SELECTOR_PATTERN) {
    selectorPattern = process.env.SELECTOR_PATTERN;
  }
  const namespaces = fs
    .readdirSync(fixturesDir, { withFileTypes: true })
    .filter((entry) => entry.isDirectory());

  const failedFixtures: FailedFixture[] = [];
  const successfulFixtures: string[] = [];

  for (const namespace of namespaces) {
    const namespaceDir: string = path.join(fixturesDir, namespace.name);
    console.log(`Testing namespace ${namespaceDir}`);
    const connectors = fs
      .readdirSync(namespaceDir, { withFileTypes: true })
      .filter((entry) => entry.isDirectory());

    for (const connector of connectors) {
      const globalConfig: GlobalConfig = {};
      if (!minimatch(`${namespace.name}/${connector.name}`, selectorPattern)) {
        console.log(`Skipping connector ${connector.name}`);
        continue;
      }
      const connectorDir: string = path.join(namespaceDir, connector.name);
      console.log(`Testing connector ${connector.name}`);
      let testConfig: TestConfig = {
        hubID: "",
        port: 8083,
        envs: undefined,
        setupComposeFile: undefined,
        runCloudTests: false,
      };
      const testConfigPath: string = path.join(
        connectorDir,
        "tests",
        "test-config.json",
      );
      try {
        if (!fs.existsSync(testConfigPath)) {
          console.error(`No test-config.json found for ${connector.name}`);
          failedFixtures.push({
            name: connector.name,
            error: `No test-config.json found for ${connector.name}`,
          });
          continue;
        }
        testConfig = JSON.parse(fs.readFileSync(testConfigPath, "utf8"));
        validateTestConfig(testConfig);

        await supergraph_init(PROJECT_DIRECTORY, false, ddn());
        let hubID = testConfig.hubID;
        if (!process.env.CONNECTOR_VERSION) {
          console.error(
            `CONNECTOR_VERSION environment variable not set. Please set it to the version of the connector to test.`,
          );
          failedFixtures.push({
            name: connector.name,
            error: `CONNECTOR_VERSION environment variable not set. Please set it to the version of the connector to test.`,
          });
          continue;
        }
        hubID = `${hubID}:${process.env.CONNECTOR_VERSION}`;
        await connector_init(PROJECT_DIRECTORY, ddn(), {
          connectorName: connector.name,
          hubID: hubID,
          port: testConfig.port || 8083,
          composeFile: "compose.yaml",
          envs: testConfig.envs || [],
        });
        if (testConfig.setupComposeFile) {
          await runCommand(
            "docker",
            [
              "compose",
              "-f",
              path.join(
                path.dirname(testConfigPath),
                testConfig.setupComposeFile,
              ),
              "up",
              "--build",
              "-d",
              "--wait",
            ],
            {
              env: {
                ...process.env,
                CONNECTOR_CONTEXT_DIR: path.join(
                  PROJECT_DIRECTORY,
                  "app",
                  "connector",
                  connector.name,
                ),
              },
            },
          );
          await sleep(10000);
        }
        await connector_introspect(PROJECT_DIRECTORY, ddn(), connector.name);
        await track_all_models(PROJECT_DIRECTORY, ddn(), connector.name);
        await track_all_commands(PROJECT_DIRECTORY, ddn(), connector.name);
        await track_all_relationships(PROJECT_DIRECTORY, ddn(), connector.name);
        await supergraph_build_local(PROJECT_DIRECTORY, ddn());
        await run_docker_start_detached(PROJECT_DIRECTORY, ddn());
        await sleep(10000);
        await run_local_tests(path.dirname(testConfigPath));
        // Run cloud tests
        try {
          if (testConfig.runCloudTests) {
            const projectName = await project_init(PROJECT_DIRECTORY, ddn());
            globalConfig.projectName = projectName;

            const pat = await printPAT(ddn());
            const projectId = await get_project_id(pat, projectName);
            const buildUrl = await supergraph_build_create(
              PROJECT_DIRECTORY,
              ddn(),
            );
            await run_cloud_tests(
              path.dirname(testConfigPath),
              ddn(),
              buildUrl,
              projectId,
            );
          }
        } catch (err) {
          console.error(
            `Error testing fixture ${connector.name} in cloud: ${err}`,
          );
          failedFixtures.push({
            name: connector.name,
            error: err,
            isCloud: true,
          });
          continue;
        }

        successfulFixtures.push(connector.name);
      } catch (e) {
        console.error(`Error testing fixture ${connector.name}: ${e}`);
        failedFixtures.push({
          name: connector.name,
          error: e,
        });
      } finally {
        try {
          if (testConfig.hubID) {
            await docker_compose_teardown(PROJECT_DIRECTORY);
          }
        } catch (err) {
          console.error(
            `Error tearing down local dc ${connector.name}: ${err}`,
          );
        }

        try {
          if (testConfig.setupComposeFile) {
            await runCommand(
              "docker",
              [
                "compose",
                "-f",
                path.join(
                  path.dirname(testConfigPath),
                  testConfig.setupComposeFile,
                ),
                "down",
                "-v",
              ],
              {
                env: {
                  ...process.env,
                  CONNECTOR_CONTEXT_DIR: path.join(
                    PROJECT_DIRECTORY,
                    "app",
                    "connector",
                    connector.name,
                  ),
                },
              },
            );
          }
        } catch (err) {
          console.error(
            `Error tearing down setup dc ${connector.name}: ${err}`,
          );
        }

        try {
          if (globalConfig.projectName) {
            await project_delete(globalConfig.projectName, PROJECT_DIRECTORY);
          }
        } catch (err) {
          console.error(
            `Error tearing down cloud project ${connector.name}: ${err}`,
          );
        }

        clear_project_dir();
      }
    }
  }

  if (successfulFixtures.length > 0) {
    console.log("Successful fixtures: ", successfulFixtures);
  }

  if (failedFixtures.length > 0) {
    console.error("Failed fixtures: ", failedFixtures);
    throw new Error(`One or more tests failed`);
  }
}
