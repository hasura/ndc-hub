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
} from "./utils.js";

interface TestConfig {
  hubID: string;
  port?: number;
  envs?: string[];
  setupComposeFile?: string;
  runCloudTests: boolean;
}

interface GlobalConfig {
  projectName: string;
}

interface FailedFixture {
  name: string;
  error: Error | string | unknown;
  isCloud?: boolean;
}

interface TestModule {
  setup: (projectDir: string, ddnCmd: string, config: GlobalConfig) => Promise<void>;
  test_local: (fixtureDir: string, config: GlobalConfig) => Promise<void>;
  test_cloud?: (projectDir: string, ddnCmd: string, fixtureDir: string, config: GlobalConfig) => Promise<void>;
  teardown: (projectDir: string, config: GlobalConfig) => Promise<void>;
}

clear_project_dir();

await setupDDNCLI();

await login();
await run_fixtures();


async function login(): Promise<void> {
  if (process.env.HASURA_DDN_PAT) {
    await runCommand(ddn(), [
      "auth",
      "login",
      "--pat",
      process.env.HASURA_DDN_PAT,
    ]);
  }
}

async function run_fixtures(
  fixturesDir: string = path.resolve(CURRENT_DIRECTORY, "..", "..", "registry")
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
      if (!minimatch(connector.name, selectorPattern)) {
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
      const testConfigPath: string = path.join(connectorDir, "tests", "test-config.json");
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

        await supergraph_init(PROJECT_DIRECTORY, false, ddn());
        await connector_init(PROJECT_DIRECTORY, ddn(), {
          connectorName: connector.name,
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
          if (testConfig.setupComposeFile) {
            await runCommand("docker", [
              "compose",
              "-f",
              path.join(path.dirname(testConfigPath), testConfig.setupComposeFile),
              "down",
              "-v",
            ], {
              env: {
                ...process.env,
                CONNECTOR_CONTEXT_DIR: path.join(PROJECT_DIRECTORY, "app", "connector", connector.name),
              },
            });
          }
        } catch (err) {
          console.error(`Error tearing down fixture ${connector.name}: ${err}`);
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

function clear_project_dir(dir: string = PROJECT_DIRECTORY): void {
  fs.rmSync(dir, { recursive: true, force: true });
  fs.mkdirSync(dir, { recursive: true });
}

function pathToFileURL(filepath: string): string {
  let normalizedPath = filepath.replace(/\\/g, "/");
  if (process.platform === "win32") {
    normalizedPath = "/" + normalizedPath;
  }
  if (!normalizedPath.startsWith("/")) {
    normalizedPath = "/" + normalizedPath;
  }
  return `file://${normalizedPath}`;
}
