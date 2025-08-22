#!/usr/bin/env bun

import { minimatch } from 'minimatch';
import { readdir, readFile } from 'fs/promises';
import { join } from 'path';
import { spawn } from 'child_process';
import { promisify } from 'util';

export interface ConnectorVersionOverride {
  name: string;
  version: string;
}


interface ConnectorTestConfig {
  connector_name: string;
  connector_version: string;
  test_config_file_path: string;
  hub_id: string;
  port?: number;
  setup_compose_file_path?: string;
  envs?: string[];
  ddn_workspace_enabled?: boolean;
  ddn_workspace_envs?: string[];
}

interface ConnectorSpec {
  publisher: string;
  name: string;
  version: string;
  hubId: string;
}

function parseConnectorSpec(spec: string): ConnectorSpec | null {
  const match = spec.match(/^([^/]+)\/([^:]+):(.+)$/);
  if (!match) {
    return null;
  }
  
  const [, publisher, name, version] = match;
  return {
    publisher,
    name,
    version,
    hubId: `${publisher}/${name}`
  };
}

const execAsync = promisify(spawn);

async function findConnectorsToTest(pattern: string = "*", versionPattern?: string): Promise<ConnectorTestConfig[]> {
  const repoRoot = process.env.NDC_HUB_GIT_REPO_FILE_PATH || process.cwd();
  const registryPath = join(repoRoot, 'registry');
  
  console.log(`🔍 Searching for connectors in: ${registryPath}`);
  
  const connectors: ConnectorTestConfig[] = [];
  
  try {
    const publishers = await readdir(registryPath);
    
    for (const publisher of publishers) {
      const publisherPath = join(registryPath, publisher);
      
      try {
        const connectorNames = await readdir(publisherPath);
        
        for (const connectorName of connectorNames) {
          if (!minimatch(connectorName, pattern)) {
            continue;
          }
          
          const connectorPath = join(publisherPath, connectorName);
          
          try {
            const versions = await readdir(connectorPath);
            
            for (const version of versions) {
              // Filter by version pattern if provided
              if (versionPattern && !minimatch(version, versionPattern)) {
                continue;
              }
              
              const versionPath = join(connectorPath, version);
              const testConfigPath = join(versionPath, 'test-config.json');
              const metadataPath = join(versionPath, 'metadata.json');
              
              try {
                // Check if test-config.json exists
                const testConfigExists = await readFile(testConfigPath, 'utf8').then(() => true).catch(() => false);
                
                if (!testConfigExists) {
                  continue;
                }
                
                // Read test config
                const testConfig = JSON.parse(await readFile(testConfigPath, 'utf8'));

                // Check if DDN workspace testing is enabled
                const ddnWorkspaceConfig = testConfig.ddn_workspace;
                if (!ddnWorkspaceConfig || !ddnWorkspaceConfig.enabled) {
                  console.log(`⏭️  Skipping ${connectorName}:${version} (ddn_workspace not enabled)`);
                  continue;
                } else {
                  console.log("DDN workspace config:", ddnWorkspaceConfig);
                }
                connectors.push({
                  connector_name: connectorName,
                  connector_version: version,
                  test_config_file_path: testConfigPath,
                  hub_id: testConfig.hub_id || `${publisher}/${connectorName}`,
                  port: testConfig.port || 8083,
                  setup_compose_file_path: testConfig.setup_compose_file_path,
                  envs: testConfig.envs || [],
                  ddn_workspace_enabled: ddnWorkspaceConfig.enabled,
                  ddn_workspace_envs: ddnWorkspaceConfig.envs || []
                });
                
                console.log(`✅ Found connector: ${connectorName}:${version}`);
                
              } catch (error) {
                console.log(`⚠️  Error processing ${connectorName}:${version} - ${error}`);
              }
            }
          } catch (error) {
            console.log(`⚠️  Error reading connector versions for ${connectorName} - ${error}`);
          }
        }
      } catch (error) {
        console.log(`⚠️  Error reading connectors for publisher ${publisher} - ${error}`);
      }
    }
  } catch (error) {
    console.error(`❌ Error reading registry directory: ${error}`);
    throw error;
  }
  
  return connectors;
}

async function runDockerCommand(command: string[], options: { cwd?: string } = {}): Promise<void> {
  return new Promise((resolve, reject) => {
    console.log(`🐳 Running: docker ${command.join(' ')}`);
    
    const childProcess = spawn('docker', command, {
      stdio: 'inherit',
      cwd: options.cwd || process.cwd()
    });
    
    childProcess.on('close', (code) => {
      if (code === 0) {
        resolve();
      } else {
        reject(new Error(`Docker command failed with exit code ${code}`));
      }
    });
    
    childProcess.on('error', (error) => {
      reject(error);
    });
  });
}

async function buildDDNWorkspace(customVersions?: ConnectorVersionOverride[]): Promise<void> {
  console.log('🏗️  Building DDN Workspace...');
  
  const repoRoot = process.env.NDC_HUB_GIT_REPO_FILE_PATH || process.cwd();
  const workspacePath = join(repoRoot, 'ddn-workspace');
  
  const buildArgs: string[] = [];
  
  if (customVersions && customVersions.length > 0) {
    console.log('📝 Using custom connector versions:', customVersions);
    
    for (const override of customVersions) {
      const argName = `${override.name.toUpperCase().replace('-', '_')}_VERSION`;
      buildArgs.push('--build-arg', `${argName}=${override.version}`);
    }
  }
  
  const dockerCommand = [
    'build',
    '-t', 'ddn-workspace:test',
    '-f', 'Dockerfile',
    '--no-cache',
    ...buildArgs,
    '.'
  ];
  
  await runDockerCommand(dockerCommand, { cwd: workspacePath });
  console.log('✅ DDN Workspace built successfully');
}

async function testConnector(connector: ConnectorTestConfig): Promise<void> {
  console.log(`🧪 Testing connector: ${connector.connector_name}:${connector.connector_version}`);
  
  const networkName = 'ddn-test-network';
  const containerName = `ddn-workspace-${connector.connector_name}`;
  
  try {
    // Create network
    await runDockerCommand(['network', 'create', networkName]).catch(() => {
      // Network might already exist, ignore error
    });

    // Track compose network name for DDN workspace to join
    let composeNetworkName: string | null = null;

    // Start setup services if needed
    if (connector.setup_compose_file_path) {
      console.log(`🐳 Starting setup services: ${connector.setup_compose_file_path}`);

      const testConfigDir = join(connector.test_config_file_path, '..');
      const composePath = join(testConfigDir, connector.setup_compose_file_path);
      
      await runDockerCommand([
        'compose',
        '-f', composePath,
        '--project-name', `setup-${connector.connector_name}`,
        'up', '-d', '--build', '--wait'
      ]);

      // Verify all services are healthy
      console.log('🏥 Verifying all services are healthy...');
      const maxRetries = 30; // 30 seconds max wait
      let retries = 0;

      while (retries < maxRetries) {
        try {
          const healthResult = await new Promise<string>((resolve, reject) => {
            const childProcess = spawn('docker', [
              'compose',
              '-f', composePath,
              '--project-name', `setup-${connector.connector_name}`,
              'ps', '--format', 'json'
            ], { stdio: 'pipe' });

            let output = '';
            childProcess.stdout.on('data', (data) => {
              output += data.toString();
            });

            childProcess.on('close', (code) => {
              if (code === 0) {
                resolve(output);
              } else {
                reject(new Error(`Docker command failed with exit code ${code}`));
              }
            });
          });

          const containers = healthResult.trim().split('\n')
            .filter((line: string) => line.length > 0)
            .map((line: string) => JSON.parse(line));

          const allHealthy = containers.every((container: any) =>
            container.Health === 'healthy' || container.Health === '' // No health check defined
          );

          if (allHealthy) {
            console.log('✅ All services are healthy');
            break;
          } else {
            const unhealthyServices = containers
              .filter((c: any) => c.Health && c.Health !== 'healthy')
              .map((c: any) => `${c.Service}(${c.Health})`)
              .join(', ');
            console.log(`⏳ Waiting for services to be healthy: ${unhealthyServices}`);
          }
        } catch (error) {
          console.log(`⏳ Waiting for services to be ready... (${retries + 1}/${maxRetries})`);
        }

        await new Promise(resolve => setTimeout(resolve, 1000)); // Wait 1 second
        retries++;
      }

      if (retries >= maxRetries) {
        console.log('⚠️ Warning: Some services may not be fully healthy, continuing anyway...');
      }

      // Store compose network name for DDN workspace container to join
      composeNetworkName = `setup-${connector.connector_name}_default`;
      console.log(`📡 Will connect DDN workspace to compose network: ${composeNetworkName}`);
    }
    
    // Build environment variables
    const envVars: string[] = [];
    
    if (process.env.HASURA_DDN_PAT) {
      envVars.push('-e', `HASURA_DDN_PAT=${process.env.HASURA_DDN_PAT}`);
    }
    
    if (process.env.DDN_WORKSPACE_ACCESS_TOKEN) {
      envVars.push('-e', `DDN_WORKSPACE_ACCESS_TOKEN=${process.env.DDN_WORKSPACE_ACCESS_TOKEN}`);
    }
    
    // Add connector-specific environment variables
    // Priority: DDN workspace envs > GitHub secrets > regular envs > empty
    let envsToUse: string[] = [];
    let envSource = '';

    if (connector.ddn_workspace_enabled && connector.ddn_workspace_envs && connector.ddn_workspace_envs.length > 0) {
      // Use explicitly configured DDN workspace environment variables
      envsToUse = connector.ddn_workspace_envs;
      envSource = 'DDN workspace configuration';
    } else if (connector.ddn_workspace_enabled) {
      // Try to read from GitHub secrets if DDN workspace is enabled but no envs configured
      const secretKey = `${connector.connector_name.toUpperCase().replace(/-/g, '_')}_CONFIG_OPTIONS_ENV`;
      const secretValue = process.env[secretKey];

      if (secretValue) {
        try {
          const secretEnvs = JSON.parse(secretValue);
          if (Array.isArray(secretEnvs)) {
            envsToUse = secretEnvs;
            envSource = `GitHub secret (${secretKey})`;
          } else {
            console.log(`⚠️ Warning: ${secretKey} is not a valid JSON array, falling back to regular envs`);
            envsToUse = connector.envs || [];
            envSource = 'regular configuration (secret invalid)';
          }
        } catch (error) {
          console.log(`⚠️ Warning: Failed to parse ${secretKey} as JSON, falling back to regular envs`);
          envsToUse = connector.envs || [];
          envSource = 'regular configuration (secret parse error)';
        }
      } else {
        // No secret found, use regular envs or empty
        envsToUse = connector.envs || [];
        envSource = envsToUse.length > 0 ? 'regular configuration' : 'empty (no configuration found)';
      }
    } else {
      // DDN workspace not enabled, use regular envs
      envsToUse = connector.envs || [];
      envSource = 'regular configuration (DDN workspace disabled)';
    }

    console.log(`📝 Using environment variables from: ${envSource} for ${connector.connector_name}`);
    if (envsToUse.length > 0) {
      console.log(`📋 Environment variables: ${envsToUse.map(env => env.split('=')[0]).join(', ')}`);
    } else {
      console.log(`📋 No environment variables configured`);
    }

    for (const env of envsToUse) {
      // Expand environment variables in the format $VAR_NAME
      const expandedEnv = env.replace(/\$([A-Z_][A-Z0-9_]*)/g, (match, varName) => {
        const value = process.env[varName];
        if (value === undefined) {
          console.log(`⚠️ Warning: Environment variable ${varName} not found, keeping ${match}`);
          return match;
        }
        console.log(`🔄 Expanding ${match} to ${value}`);
        return value;
      });
      envVars.push('-e', expandedEnv);
    }
    
    // Start DDN workspace container
    // Use compose network if available, otherwise use DDN test network
    const targetNetwork = composeNetworkName || networkName;
    console.log(`🔗 Connecting DDN workspace to network: ${targetNetwork}`);

    await runDockerCommand([
      'run', '-d',
      '--name', containerName,
      '--network', targetNetwork,
      '--privileged',
      '--entrypoint', '',
      ...envVars,
      'ddn-workspace:test',
      'bash', '-c', `
        dockerd --host=unix:///var/run/docker.sock --host=tcp://0.0.0.0:2376 &
        while ! docker info >/dev/null 2>&1; do sleep 1; done
        echo 'Docker daemon started successfully'
        
        export PATH="$HOME/.local/bin:$PATH"
      `
    ]);
    
    // Wait for container to be ready
    console.log('⏳ Waiting for DDN workspace to initialize...');
    await new Promise(resolve => setTimeout(resolve, 20000));
    
    // Run introspection test
    console.log('🔍 Running connector introspection test...');
    
    const validConnectorName = connector.connector_name.replace('-', '_');
    
    await runDockerCommand([
      'exec', containerName,
      'bash', '-c', `
        set -e
        export PATH="$HOME/.local/bin:$PATH"
        
        echo '🚀 Starting DDN workspace introspection test'
        
        mkdir -p /tmp/test-project
        cd /tmp/test-project

        echo 'Authenticating with DDN...'
        ddn auth login --access-token "$DDN_WORKSPACE_ACCESS_TOKEN" || echo 'DDN login failed, continuing...'
        
        echo '1️⃣ Initializing supergraph...'
        ddn supergraph init .


        
        echo "Using valid connector name: ${validConnectorName}"
        
        echo '2️⃣ Initializing connector: ${validConnectorName}'
        echo "Command: ddn connector init ${validConnectorName} --hub-connector ${connector.hub_id}:${connector.connector_version} --configure-port ${connector.port}"
        ddn connector init ${validConnectorName} \\
          --hub-connector ${connector.hub_id}:${connector.connector_version} \\
          --configure-port ${connector.port}
        
        echo '3️⃣ Running connector introspection...'
        ddn connector introspect ${validConnectorName}
        
        echo '✅ DDN workspace introspection test PASSED'
        echo 'Introspected files:'
        find app/connector/${validConnectorName} -type f | head -10
      `
    ]);
    
    console.log(`✅ Connector test completed successfully: ${connector.connector_name}`);
    
  } finally {
    // Cleanup
    await runDockerCommand(['stop', containerName]).catch(() => {});
    await runDockerCommand(['rm', containerName]).catch(() => {});
    
    if (connector.setup_compose_file_path) {
      const testConfigDir = join(connector.test_config_file_path, '..');
      const composePath = join(testConfigDir, connector.setup_compose_file_path);
      
      await runDockerCommand([
        'compose',
        '-f', composePath,
        '--project-name', `setup-${connector.connector_name}`,
        'down', '-v'
      ]).catch(() => {});
    }
  }
}

async function findSpecificConnector(connectorSpec: string): Promise<ConnectorTestConfig[]> {
  const spec = parseConnectorSpec(connectorSpec);
  if (!spec) {
    throw new Error(`Invalid connector specification: ${connectorSpec}. Expected format: publisher/name:version`);
  }
  
  const repoRoot = process.env.NDC_HUB_GIT_REPO_FILE_PATH || join(process.cwd(), '../..');
  const connectorPath = join(repoRoot, 'registry', spec.publisher, spec.name);
  const testConfigPath = join(connectorPath, 'tests', 'test-config.json');
  const versionPath = join(connectorPath, spec.version);
  const metadataPath = join(versionPath, 'metadata.json');
  
  console.log(`🔍 Looking for specific connector: ${connectorSpec}`);
  
  try {
    // Check if test-config.json exists at connector level
    const testConfigExists = await readFile(testConfigPath, 'utf8').then(() => true).catch(() => false);
    
    if (!testConfigExists) {
      throw new Error(`test-config.json not found for ${connectorSpec} at ${testConfigPath}`);
    }
    
    // Read test config
    const testConfig = JSON.parse(await readFile(testConfigPath, 'utf8'));

    // Check if DDN workspace testing is enabled
    const ddnWorkspaceConfig = testConfig.ddn_workspace;
    if (!ddnWorkspaceConfig || !ddnWorkspaceConfig.enabled) {
      console.log(`⚠️  Warning: ${connectorSpec} does not have ddn_workspace enabled - skipping`);
      return [];
    }

    const connector: ConnectorTestConfig = {
      connector_name: spec.name,
      connector_version: spec.version,
      test_config_file_path: testConfigPath,
      hub_id: testConfig.hub_id || spec.hubId,
      port: testConfig.port || 8083,
      setup_compose_file_path: testConfig.setup_compose_file_path,
      envs: testConfig.envs || [],
      ddn_workspace_enabled: ddnWorkspaceConfig.enabled,
      ddn_workspace_envs: ddnWorkspaceConfig.envs || []
    };
    
    console.log(`✅ Found connector: ${spec.name}:${spec.version}`);
    return [connector];
    
  } catch (error) {
    console.error(`❌ Error finding connector ${connectorSpec}: ${error}`);
    throw error;
  }
}

async function findConnectorsFromVersionsFile(): Promise<ConnectorTestConfig[]> {
  console.log('📋 Loading connectors from connector-versions.json');

  const repoRoot = process.env.NDC_HUB_GIT_REPO_FILE_PATH || join(process.cwd(), '../..');
  const versionsFilePath = join(repoRoot, 'ddn-workspace', 'connector-versions.json');

  try {
    const versionsContent = await readFile(versionsFilePath, 'utf8');
    const versions = JSON.parse(versionsContent);

    const connectors: ConnectorTestConfig[] = [];

    for (const [connectorName, version] of Object.entries(versions)) {
      const connectorSpec = `hasura/${connectorName}:${version}`;
      console.log(`🔍 Processing connector from versions file: ${connectorSpec}`);

      try {
        const connectorConfigs = await findSpecificConnector(connectorSpec);
        connectors.push(...connectorConfigs);
      } catch (error) {
        console.log(`⚠️  Warning: Failed to load ${connectorSpec}: ${error}`);
      }
    }

    return connectors;
  } catch (error) {
    throw new Error(`Failed to load connector-versions.json: ${error}`);
  }
}

async function findSpecificConnectors(connectorSpecs: string[]): Promise<ConnectorTestConfig[]> {
  console.log(`🔍 Loading specific connectors: ${connectorSpecs.join(', ')}`);

  const connectors: ConnectorTestConfig[] = [];

  for (const spec of connectorSpecs) {
    try {
      const connectorConfigs = await findSpecificConnector(spec.trim());
      connectors.push(...connectorConfigs);
    } catch (error) {
      console.log(`⚠️  Warning: Failed to load ${spec}: ${error}`);
    }
  }

  return connectors;
}

export async function runDDNWorkspaceTestSuite(input?: string, customVersions?: ConnectorVersionOverride[]): Promise<void> {
  console.log('🚀 Starting DDN Workspace Test Suite');

  let connectors: ConnectorTestConfig[];

  if (!input || input === "*") {
    // Mode 1: Test all connectors from connector-versions.json
    connectors = await findConnectorsFromVersionsFile();
  } else {
    // Mode 2: Test specific connector(s)
    const connectorSpecs = input.split(',').map(s => s.trim()).filter(s => s.length > 0);
    connectors = await findSpecificConnectors(connectorSpecs);
  }

  if (connectors.length === 0) {
    console.log('⚠️  No connectors found to test');
    return;
  }

  console.log(`📋 Found ${connectors.length} connectors to test`);
  // Test each connector
  for (const connector of connectors) {
    try {
      await testConnector(connector);
      console.log(`✅ Connector test completed successfully: ${connector.connector_name}`);
    } catch (error) {
      console.error(`❌ Test failed for ${connector.connector_name}:${connector.connector_version}`, error);
      throw error;
    }
  }
  
  console.log('🎉 All DDN workspace tests completed successfully!');
}
