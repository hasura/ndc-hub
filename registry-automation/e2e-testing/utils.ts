import { spawn, type ChildProcess } from "child_process";
import { fileURLToPath } from "url";
import path from "path";
import yaml from "yaml";
import fs from "fs";
import os from "os";
import assert from "node:assert/strict";
import https from "https";
import axios, { type AxiosInstance } from "axios";

export const CURRENT_DIRECTORY: string = path.dirname(
  fileURLToPath(import.meta.url),
);
export const PROJECT_DIRECTORY: string = path.join(
  CURRENT_DIRECTORY,
  "project",
);
export const FIXTURES_DIRECTORY: string =
  process.env.FIXTURES_DIRECTORY || path.join(CURRENT_DIRECTORY, "fixtures");
export const ENGINE_PORT: string = process.env.ENGINE_PORT || "3280";
export const PROMPTQL_PORT: string = process.env.PROMPTQL_PORT || "3282";
export const IS_CLOUD_TEST_ENABLED: boolean =
  process.env.ENABLE_CLOUD_TESTS === "true";
export const AUTH_ENDPOINT: string =
  process.env.AUTH_ENDPOINT || "https://auth.pro.hasura.io";
export const DATA_ENDPOINT: string =
  process.env.DATA_ENDPOINT || "https://data.pro.hasura.io";
export const PROMPTQL_ENDPOINT: string | undefined =
  process.env.PROMPTQL_ENDPOINT;

export const REASONABLE_PAUSE: number = 5;

interface CommandResult {
  code: number;
  stdout: string;
  stderr: string;
}

interface CommandOptions extends Record<string, any> {
  suppressOutput?: boolean;
  shell?: boolean;
  cwd?: string;
  env?: NodeJS.ProcessEnv;
}

interface ConnectorInitOptions {
  connectorName: string;
  hubID: string;
  port: number;
  composeFile: string;
  envs: string[];
}

export function sleep(ms: number): Promise<void> {
  console.log(`Pausing for ${ms}ms`);
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

function createAxiosClient(
  url: string,
  headers: Record<string, string>,
): AxiosInstance {
  return axios.create({
    baseURL: url,
    headers: headers,
  });
}

export async function runCommand(
  command: string | Promise<string>,
  args: string[] = [],
  options: CommandOptions = {},
): Promise<CommandResult> {
  command = await command;
  return new Promise((resolve, reject) => {
    const stdoutData: Buffer[] = [];
    const stderrData: Buffer[] = [];

    if (!("shell" in options)) {
      options.shell = true;
    }

    console.log(`Running command "${command} ${args.join(" ")}"`);
    const cmd: ChildProcess = spawn(command, args, options);

    cmd.stdout?.on("data", (data: Buffer) => {
      if (!options.suppressOutput) {
        process.stdout.write(data.toString());
      }
      stdoutData.push(data);
    });

    cmd.stderr?.on("data", (data: Buffer) => {
      if (!options.suppressOutput) {
        process.stderr.write(data.toString());
      }
      stderrData.push(data);
    });

    cmd.on("close", (code: number | null) => {
      const fullStdout = Buffer.concat(stdoutData).toString();
      const fullStderr = Buffer.concat(stderrData).toString();

      if (code !== 0) {
        console.error(
          `Command "${command} ${args.join(" ")}" failed with code ${code}`,
        );
        reject(new Error(fullStderr));
        return;
      }

      resolve({
        code: code || 0,
        stdout: fullStdout,
        stderr: fullStderr,
      });
    });

    cmd.on("error", (err: Error) => {
      reject(err);
    });
  });
}

export async function supergraph_init(
  dir: string = PROJECT_DIRECTORY,
  promptql: boolean = false,
  ddnCmd: string = ddn(),
): Promise<string | undefined> {
  const args: string[] = ["supergraph", "init", ".", "--out", "json"];
  if (promptql) {
    args.push("--with-promptql");
    args.push("--subgraph-naming-convention");
    args.push("graphql");
  }
  const res = await runCommand(ddnCmd, args, {
    cwd: dir,
  });
  await update_context_yaml(dir, ddnCmd);

  if (res.stdout) {
    try {
      const data = JSON.parse(res.stdout);
      return data.project;
    } catch (e) {
      console.log(e);
      return undefined;
    }
  }
  return undefined;
}

export function validConnectorName(connectorName: string): string {
  return connectorName.replace(/-/g, '_');
}

export async function connector_init(
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string = ddn(),
  options: ConnectorInitOptions,
): Promise<void> {
  const args: string[] = [
    "connector",
    "init",
    `"${options.connectorName}"`,
    "--hub-connector",
    `"${options.hubID}"`,
    "--configure-port",
    `"${options.port}"`,
    "--add-to-compose-file",
    `"${options.composeFile}"`,
  ];
  // Check for environment variables that end with _CONFIG_OPTIONS_ENV and start with the connector name
  for (const key of Object.keys(process.env).filter(k => k.endsWith('_CONFIG_OPTIONS_ENV') && k.startsWith(options.connectorName.toUpperCase()))) {
    const value = process.env[key];
    if (value) {
      try {
        const options = JSON.parse(value); // Try parsing as JSON array
        for (const env of options) {
          args.push("--add-env", `"${env}"`);
        }
      } catch (err) {
        console.error(`Invalid format for ${key}:`, err);
      }
    }
  }
  if (options.envs) {
    for (const env of options.envs) {
      args.push("--add-env", `"${env}"`);
    }
  }
  await runCommand(ddnCmd, args, {
    cwd: dir,
  });
}

export async function connector_introspect(
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string = ddn(),
  connectorName: string,
): Promise<void> {
  const args = ["connector", "introspect", `"${connectorName}"`];
  await runCommand(ddnCmd, args, {
    cwd: dir,
  });
}

export async function track_all_models(
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string = ddn(),
  connectorLinkName: string,
): Promise<void> {
  await track(dir, ddnCmd, "model", connectorLinkName, "*");
}

export async function track_all_commands(
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string = ddn(),
  connectorLinkName: string,
): Promise<void> {
  await track(dir, ddnCmd, "command", connectorLinkName, "*");
}

export async function track_all_relationships(
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string = ddn(),
  connectorLinkName: string,
): Promise<void> {
  await track(dir, ddnCmd, "relationship", connectorLinkName, "*");
}

export async function supergraph_build_local(
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string = ddn(),
): Promise<void> {
  const args = ["supergraph", "build", "local"];
  await runCommand(ddnCmd, args, {
    cwd: dir,
  });
}

export async function run_docker_start_detached(
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string = ddn(),
): Promise<void> {
  const args = ["run", "docker-start", "--", "-d", "--wait"];
  await runCommand(ddnCmd, args, {
    cwd: dir,
  });
}

export async function docker_compose_teardown(
  dir: string = PROJECT_DIRECTORY,
): Promise<void> {
  const args = ["compose", "down", "-v"];
  await runCommand("docker", args, {
    cwd: dir,
  });
}

export async function run_local_tests(
  fixtureName: string,
  snapshotDir: string,
): Promise<void> {
  const client = createLocalGQLClient();
  await run_tests(fixtureName, snapshotDir, client);
}

export async function run_cloud_tests(
  fixtureName: string,
  snapshotDir: string,
  ddnCmd: string = ddn(),
  buildURL: string,
  projectId: string,
): Promise<void> {
  const client = await createCloudGQLClient(
    buildURL,
    await printPAT(ddnCmd),
    projectId,
  );
  await run_tests(fixtureName, snapshotDir, client);
}

async function run_tests(
  fixtureName: string,
  snapshotDir: string,
  client: AxiosInstance,
): Promise<void> {
  if (!fs.existsSync(snapshotDir)) {
    console.log(
      `No snapshots found. Skipping tests for shapshotDir ${snapshotDir}`,
    );
    return;
  }
  const entries = fs.readdirSync(snapshotDir, { withFileTypes: true });
  const directories = entries
    .filter((entry) => entry.isDirectory())
    .map((dir) => dir.name);

  let testFailure = false;
  for (const dir of directories) {
    try {
      console.log(`Testing snapshot "${dir}" of fixture "${fixtureName}"`);
      await runTest(client, path.join(snapshotDir, dir));
      console.log(`Passed snapshot "${dir}" of fixture "${fixtureName}"`);
    } catch (e) {
      testFailure = true;
      console.error(
        `Snapshot "${dir}" of fixture "${fixtureName}" failed. Error: `,
        e,
      );
    }
  }
  if (testFailure) {
    throw new Error(`Test failure for fixture "${fixtureName}"`);
  }
}

export async function subgraph_create(
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string | Promise<string> = ddn(),
  subgraphName: string,
  projectName: string,
): Promise<void> {
  if (!projectName || !subgraphName) return;

  const args = ["project", "subgraph", "create"];
  args.push(subgraphName);
  args.push("--project");
  args.push(projectName);

  await runCommand(ddnCmd, args, {
    cwd: dir,
  });
}

export async function project_create(
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string | Promise<string> = ddn(),
  projectName: string,
): Promise<void> {
  const args = ["project", "create"];
  if (projectName) {
    args.push(projectName);
  }

  await runCommand(ddnCmd, args, {
    cwd: dir,
  });
}

export async function project_init(
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string | Promise<string> = ddn(),
): Promise<string> {
  const args = ["project", "init", "--out", "json"];

  const res = await runCommand(ddnCmd, args, {
    cwd: dir,
  });

  const data = JSON.parse(res.stdout);
  return data.project;
}

export async function readProjectNameFromContext(
  dir: string = PROJECT_DIRECTORY,
): Promise<string> {
  const contextPath = path.join(dir, ".hasura", "context.yaml");
  const context = yaml.parse(fs.readFileSync(contextPath, "utf-8"));
  return context.definition.contexts.default.project;
}

export async function project_delete(
  projectName: string,
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string | Promise<string> = ddn(),
): Promise<void> {
  if (!projectName) return;

  const args = ["project", "delete"];
  args.push(projectName);
  args.push("-f");

  await runCommand(ddnCmd, args, {
    cwd: dir,
  });
}

export async function supergraph_build_create(
  dir: string = PROJECT_DIRECTORY,
  ddnCmd: string | Promise<string> = ddn(),
): Promise<string> {
  const args = ["supergraph", "build", "create", "--out", "json"];
  const { stdout } = await runCommand(ddnCmd, args, {
    cwd: dir,
  });
  const output = JSON.parse(stdout);

  // build takes time to complete, 404 is thrown if you try to access the build_url immediately
  await sleep(10000);

  return output.build_url;
}

async function runTest(
  client: AxiosInstance,
  snapshotDir: string,
  requestFile: string = "request.graphql",
  responseFile: string = "response.json",
  variablesFile: string = "variables.json",
): Promise<void> {
  const query = fs.readFileSync(path.join(snapshotDir, requestFile), "utf-8");
  let variables = {};
  if (fs.existsSync(path.join(snapshotDir, variablesFile))) {
    variables = JSON.parse(
      fs.readFileSync(path.join(snapshotDir, variablesFile), "utf-8"),
    );
  }

  const response = await client.post("", {
    query: query,
    variables: variables,
  });

  const expectedResponse = JSON.parse(
    fs.readFileSync(path.join(snapshotDir, responseFile), "utf-8"),
  );
  assert.deepEqual(response.data, expectedResponse);
}

function createLocalGQLClient(
  options = {
    endpoint: `http://localhost:${ENGINE_PORT}/graphql`,
    headers: {},
  },
): AxiosInstance {
  const client = createAxiosClient(options.endpoint, options.headers);
  return client;
}

export async function promptql_local_test(dir: string): Promise<void> {
  console.log("INFO: Checking if promtpql is enabled for local project");

  const contextPath = path.join(dir, ".hasura", "context.yaml");
  const context = yaml.parse(fs.readFileSync(contextPath, "utf-8"));

  if (!context.definition.promptQL) {
    throw new Error("PromptQL is not enabled for local project");
  }

  if (!context.definition.contexts.default.project) {
    throw new Error(
      "Project is not set in context.yaml, promptql will not work properly",
    );
  }

  const compose = yaml.parse(
    fs.readFileSync(path.join(dir, "compose.yaml"), "utf-8"),
  );
  if (!compose.services["promptql-playground"]) {
    throw new Error(
      "PromptQL playground service is not enabled in compose.yaml",
    );
  }

  console.log("SUCCESS: PromptQL is enabled for local project");
  console.log(
    "INFO: Checking if PromptQL service is running locally and is healthy",
  );

  const config = await axios.get(
    `http://localhost:${PROMPTQL_PORT}/config-check`,
    {
      headers: {},
    },
  );

  if (config.status !== 200 || config.data != "OK") {
    throw new Error("PromptQL config check was not successful");
  }

  const threads = await axios.get(`http://localhost:${PROMPTQL_PORT}/threads`, {
    headers: {},
  });

  if (threads.status !== 200) {
    throw new Error("PromptQL threads check was not successful");
  }

  console.log("SUCCESS: PromptQL service is running locally and is healthy");
}

export async function promptql_cloud_test(
  pat: string,
  projectId: string,
): Promise<void> {
  if (!PROMPTQL_ENDPOINT) {
    console.log(
      "INFO: PROMPTQL_ENDPOINT is not set, skipping PromptQL cloud test",
    );
    return;
  }

  // promptql client
  const pqlClient = createAxiosClient(`${PROMPTQL_ENDPOINT}`, {
    Authorization: `pat ${pat}`,
  });

  // fetch project id
  const query = `
query GetPromptqlConfig ($projectId: String!) {
	getPromptqlConfig(projectId: $projectId) {
		promptQlEnabled
		playgroundEnabled
	}
}`;

  const res = await pqlClient.post("", {
    query,
    variables: { projectId },
  });

  console.log("INFO: Testing if PromptQL is enabled for cloud project");
  if (res.data.data.getPromptqlConfig.promptQlEnabled !== true) {
    throw new Error("PromptQL is not enabled for cloud project");
  }
  console.log("INFO: PromptQL is enabled for cloud project");

  console.log(
    "INFO: Testing if PromptQL Playground is enabled for cloud project",
  );
  if (res.data.data.getPromptqlConfig.playgroundEnabled !== true) {
    throw new Error("PromptQL Playground is not enabled for cloud project");
  }
  console.log("INFO: PromptQL Playground is enabled for cloud project");
}

export async function enable_promptql(
  dir = PROJECT_DIRECTORY,
  ddnCmd = ddn(),
): Promise<void> {
  const args = [
    "codemod",
    "enable-promptql",
    "-f",
    "--compose-file-path",
    "compose.yaml",
  ];

  await runCommand(ddnCmd, args, {
    cwd: dir,
  });
}

export async function get_project_id(
  pat: string,
  projectName: string,
): Promise<string> {
  // cps client
  const cpsClient = createAxiosClient(`${DATA_ENDPOINT}/v1/graphql`, {
    Authorization: `pat ${pat}`,
  });

  // fetch project id
  const project_query = `
  query DescribeProject($projectName: String!) {
    ddn_projects(where: { name: { _eq: $projectName } }) {
      id
    }
  }`;

  console.log("INFO: Fetching project id for cloud project");
  const res = await cpsClient.post("", {
    query: project_query,
    variables: { projectName },
  });

  const projectId = res.data.data.ddn_projects[0].id;
  return projectId;
}

async function createCloudGQLClient(
  buildURL: string,
  pat: string,
  projectId: string,
): Promise<AxiosInstance> {
  // fetch ddn token
  console.log("INFO: Fetching ddn token for cloud project");
  const res = await axios.post(
    `${AUTH_ENDPOINT}/ddn/project/token`,
    {},
    {
      headers: {
        Authorization: `pat ${pat}`,
        "x-hasura-project-id": projectId,
      },
    },
  );

  const client = createAxiosClient(buildURL, {
    "x-hasura-ddn-token": res.data.token,
  });

  return client;
}

export async function printPAT(ddnCmd = ddn()): Promise<string> {
  const args = ["auth", "print-pat"];
  const { stdout } = await runCommand(ddnCmd, args, {
    cwd: PROJECT_DIRECTORY,
    suppressOutput: true,
  });
  const pat = stdout.trim();
  return pat;
}

export async function track(
  dir = PROJECT_DIRECTORY,
  ddnCmd = ddn(),
  entityType: string,
  connectorLinkName: string,
  pattern: string,
): Promise<void> {
  const args = [
    `${entityType}`,
    "add",
    `"${connectorLinkName}"`,
    `"${pattern}"`,
  ];
  await runCommand(ddnCmd, args, {
    cwd: dir,
  });
}

async function update_context_yaml(
  dir = PROJECT_DIRECTORY,
  ddnCommand = ddn(),
) {
  const ddnCmd = await ddnCommand;
  const contextPath = path.join(dir, ".hasura", "context.yaml");
  const context = yaml.parse(fs.readFileSync(contextPath, "utf-8"));
  context.definition.scripts["docker-start"].bash = context.definition.scripts[
    "docker-start"
  ].bash.replaceAll("ddn auth", `${ddnCmd} auth`);
  context.definition.scripts["docker-start"].powershell =
    context.definition.scripts["docker-start"].powershell.replaceAll(
      "ddn auth",
      `${ddnCmd} auth`,
    );
  fs.writeFileSync(contextPath, yaml.stringify(context), "utf-8");
}

function getArchitecture(): string {
  const arch = os.arch();
  if (arch === "x64") {
    return "amd64";
  }
  return "arm64";
}

function getOs(): string {
  const platform = os.platform();
  if (platform === "win32") {
    return "windows";
  }
  return platform;
}

function getSuffix(): string {
  const platform = os.platform();
  if (platform === "win32") {
    return ".exe";
  }
  return "";
}

async function downloadBinary(
  url: string,
  destination: string,
): Promise<string> {
  if (url === undefined) {
    return Promise.reject(
      new Error("Specify either binary path or binary environment"),
    );
  } else {
    console.log(`Using ${url} to download CLI:`);

    return new Promise((resolve, reject) => {
      const file = fs.createWriteStream(destination, { mode: 0o755 });

      https
        .get(url, (response) => {
          console.log(`HTTP GET request sent to: ${url}`);
          console.log(`Response status code: ${response.statusCode}`);

          if (response.statusCode !== 200) {
            reject(
              new Error(
                `Failed to download file: ${response.statusCode} ${response.statusMessage}`,
              ),
            );
            return;
          }

          response.pipe(file);

          file.on("finish", () => {
            console.log(`Download complete. Closing file: ${destination}`);
            file.close(() => {
              console.log(`Binary downloaded and saved to: ${destination}`);
              resolve(destination);
            });
          });

          file.on("error", (err) => {
            console.error(`Error writing to file: ${err.message}`);
            fs.unlink(destination, () => reject(err));
          });
        })
        .on("error", (err) => {
          console.error(`Error during HTTP GET request: ${err.message}`);
          fs.unlink(destination, () => reject(err));
        });
    });
  }
}

let DDN_CLI_PATH: string;

export async function setupDDNCLI(): Promise<void> {
  DDN_CLI_PATH = await pathToDDNCLI();
}

export async function pathToDDNCLI(): Promise<string> {
  const cliName = `cli-ddn-${getOs()}-${getArchitecture()}${getSuffix()}`;
  if (process.env.DDN_CLI_DIRECTORY) {
    console.log(
      `Attempting to use local binary from directory: ${process.env.DDN_CLI_DIRECTORY}`,
    );
    const cliDir = path.resolve(process.env.DDN_CLI_DIRECTORY);
    const cliBinaryName = process.env.DDN_CLI_BINARY_NAME || cliName;
    const localPath = path.join(cliDir, cliBinaryName);
    if (fs.existsSync(localPath)) {
      const stats = fs.statSync(localPath);
      if (stats.isFile()) {
        console.log(`Using binary from: ${localPath}`);
        return localPath;
      } else {
        throw new Error(`Path is not a file: ${localPath}`);
      }
    } else {
      throw new Error(`Binary not found at: ${localPath}`);
    }
  } else {
    console.log("DDN_CLI_DIRECTORY not provided. Attempting to download CLI");
  }

  let downloadURL: string;

  if (process.env.DDN_CLI_DOWNLOAD_URL) {
    console.log(
      `Attempting to use download URL from: ${process.env.DDN_CLI_DOWNLOAD_URL}`,
    );
    downloadURL = process.env.DDN_CLI_DOWNLOAD_URL;
  } else {
    if (!process.env.CLI_TAG) {
      throw new Error(
        "Either CLI_TAG or DDN_CLI_DOWNLOAD_URL must be set to download the CLI",
      );
    }
    downloadURL = `https://graphql-engine-cdn.hasura.io/ddn/cli/v4/${process.env.CLI_TAG}/${cliName}`;
  }

  const cliDownloadPath = path.join(CURRENT_DIRECTORY, cliName);
  return await downloadBinary(downloadURL, cliDownloadPath);
}

export function ddn(): string {
  if (DDN_CLI_PATH) {
    return DDN_CLI_PATH;
  }
  throw new Error("DDN CLI not set up");
}

export interface GlobalConfig {
  projectName?: string;
}

export interface FailedFixture {
  name: string;
  error: Error | string | unknown;
  isCloud?: boolean;
}

export function clear_project_dir(dir: string = PROJECT_DIRECTORY): void {
  fs.rmSync(dir, { recursive: true, force: true });
  fs.mkdirSync(dir, { recursive: true });
}

export function pathToFileURL(filepath: string): string {
  let normalizedPath = filepath.replace(/\\/g, "/");
  if (process.platform === "win32") {
    normalizedPath = "/" + normalizedPath;
  }
  if (!normalizedPath.startsWith("/")) {
    normalizedPath = "/" + normalizedPath;
  }
  return `file://${normalizedPath}`;
}

export async function login(): Promise<void> {
  if (process.env.HASURA_DDN_PAT) {
    await runCommand(ddn(), [
      "auth",
      "login",
      "--pat",
      process.env.HASURA_DDN_PAT,
    ]);
  }
}
