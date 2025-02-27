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
} from "./utils.js";

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

async function run_fixtures(fixturesDir: string = FIXTURES_DIRECTORY): Promise<void> {
  let selectorPattern: string = "*";
  if (process.env.SELECTOR_PATTERN) {
    selectorPattern = process.env.SELECTOR_PATTERN;
  }
  const entries = fs.readdirSync(fixturesDir, { withFileTypes: true });
  const globalConfig = { projectName: "" };

  const directories = entries
    .filter((entry) => entry.isDirectory())
    .map((dir) => dir.name);

  const failedFixtures = [];
  const successfulFixtures = [];

  for (const dir of directories) {
    if (minimatch(dir, selectorPattern)) {
      clear_project_dir();
      let module: any;
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
              globalConfig,
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
