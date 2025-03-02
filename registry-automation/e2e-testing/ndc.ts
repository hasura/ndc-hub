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
  clear_project_dir,
  login,
  project_init,
  type GlobalConfig,
  printPAT,
  get_project_id,
  supergraph_build_create,
  run_cloud_tests,
  project_delete,
} from "./utils";

clear_project_dir();

await setupDDNCLI();

await login();
await run_fixtures();

function read_job_config(): TestJob[] {
  const testJobFile = process.env.TEST_JOB_FILE;
  if (!testJobFile) {
    throw new Error(
      `Provide TEST_JOB_FILE env var with the path to the job config json file`,
    );
  }
  return JSON.parse(fs.readFileSync(testJobFile, "utf8")) as TestJob[];
}

function read_test_config(job: TestJob): TestConfig {
  // job.test_config_file_path is relative to the repo. Use the repo root to construct Abs path
  const repoRoot = process.env.NDC_HUB_GIT_REPO_FILE_PATH;
  if (!repoRoot) {
    throw new Error(
      `NDC_HUB_GIT_REPO_FILE_PATH env var is not set. Please set it to the root of the ndc-hub repo`,
    );
  }
  const tc = JSON.parse(
    fs.readFileSync(path.join(repoRoot, job.test_config_file_path), "utf8"),
  ) as TestConfig;
  tc._testConfigDir = path.dirname(
    path.join(repoRoot, job.test_config_file_path),
  );
  return tc;
}

function get_snapshot_dir(testConfig: TestConfig): string {
  return path.join(testConfig._testConfigDir, testConfig.snapshots_dir);
}

async function run_fixtures(): Promise<void> {
  const jobs = read_job_config();
  const failedFixtures: FailedFixture[] = [];
  const successfulFixtures: string[] = [];

  for (const job of jobs) {
    const globalConfig: GlobalConfig = {};
    let testConfig: TestConfig;
    try {
      testConfig = read_test_config(job);
      validateTestConfig(testConfig);
    } catch (e) {
      console.error(
        `Error reading test config for ${job.connector_name}: ${e}`,
      );
      failedFixtures.push({
        name: job.connector_name,
        error: e,
      });
      continue;
    }

    const connectorID = `${testConfig.hub_id}:${job.connector_version}`;

    try {
      console.log(`Testing connector ${connectorID}`);

      await supergraph_init(PROJECT_DIRECTORY, false, ddn());

      await connector_init(PROJECT_DIRECTORY, ddn(), {
        connectorName: job.connector_name,
        hubID: connectorID,
        port: testConfig.port || 8083,
        composeFile: "compose.yaml",
        envs: testConfig.envs || [],
      });

      if (testConfig.setup_compose_file_path) {
        await runCommand(
          "docker",
          [
            "compose",
            "-f",
            path.join(
              path.dirname(job.test_config_file_path),
              testConfig.setup_compose_file_path,
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
                job.connector_name,
              ),
            },
          },
        );
        await sleep(10000);
      }

      await connector_introspect(PROJECT_DIRECTORY, ddn(), job.connector_name);

      await track_all_models(PROJECT_DIRECTORY, ddn(), job.connector_name);

      await track_all_commands(PROJECT_DIRECTORY, ddn(), job.connector_name);

      await track_all_relationships(
        PROJECT_DIRECTORY,
        ddn(),
        job.connector_name,
      );

      await supergraph_build_local(PROJECT_DIRECTORY, ddn());

      await run_docker_start_detached(PROJECT_DIRECTORY, ddn());
      await sleep(10000);

      await run_local_tests(connectorID, get_snapshot_dir(testConfig));
      // Run cloud tests
      try {
        if (testConfig.run_cloud_tests) {
          const projectName = await project_init(PROJECT_DIRECTORY, ddn());
          globalConfig.projectName = projectName;

          const pat = await printPAT(ddn());
          const projectId = await get_project_id(pat, projectName);
          const buildUrl = await supergraph_build_create(
            PROJECT_DIRECTORY,
            ddn(),
          );
          await run_cloud_tests(
            connectorID,
            get_snapshot_dir(testConfig),
            ddn(),
            buildUrl,
            projectId,
          );
        }
      } catch (err) {
        console.error(`Error testing fixture ${connectorID} in cloud: ${err}`);
        failedFixtures.push({
          name: connectorID,
          error: err,
          isCloud: true,
        });
        continue;
      }

      successfulFixtures.push(connectorID);
    } catch (e) {
      console.error(`Error testing fixture ${connectorID}: ${e}`);
      failedFixtures.push({
        name: connectorID,
        error: e,
      });
    } finally {
      try {
        if (testConfig.hub_id) {
          await docker_compose_teardown(PROJECT_DIRECTORY);
        }
      } catch (err) {
        console.error(`Error tearing down local dc ${connectorID}: ${err}`);
      }

      try {
        if (testConfig.setup_compose_file_path) {
          await runCommand(
            "docker",
            [
              "compose",
              "-f",
              path.join(
                testConfig._testConfigDir,
                testConfig.setup_compose_file_path,
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
                  job.connector_name,
                ),
              },
            },
          );
        }
      } catch (err) {
        console.error(`Error tearing down setup dc ${connectorID}: ${err}`);
      }

      try {
        if (globalConfig.projectName) {
          await project_delete(globalConfig.projectName, PROJECT_DIRECTORY);
        }
      } catch (err) {
        console.error(
          `Error tearing down cloud project ${connectorID}: ${err}`,
        );
      }

      clear_project_dir();
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

interface TestConfig {
  hub_id: string;
  port?: number;
  envs?: string[];
  setup_compose_file_path?: string;
  run_cloud_tests?: boolean;
  snapshots_dir: string;

  // Internal Properties
  _testConfigDir: string;
}

interface TestJob {
  namespace: string;
  connector_name: string;
  connector_version: string;
  test_config_file_path: string;
}

function validateTestConfig(config: TestConfig): void {
  if (!config.hub_id) {
    throw new Error("hub_id is required in test config");
  }
  if (!config.snapshots_dir) {
    throw new Error("snapshots_dir is required in test config");
  }
  if (!config._testConfigDir) {
    throw new Error("_testConfigDir must be set in test config");
  }
}
