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
  setup_compose_path?: string;
  envs?: string[];
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
                
                // Read metadata to check if add-to-ddn-workspace is true
                const metadataExists = await readFile(metadataPath, 'utf8').then(() => true).catch(() => false);
                
                if (metadataExists) {
                  const metadata = JSON.parse(await readFile(metadataPath, 'utf8'));
                  if (!metadata['add-to-ddn-workspace']) {
                    console.log(`⏭️  Skipping ${connectorName}:${version} (add-to-ddn-workspace not enabled)`);
                    continue;
                  }
                }
                
                // Read test config
                const testConfig = JSON.parse(await readFile(testConfigPath, 'utf8'));
                
                connectors.push({
                  connector_name: connectorName,
                  connector_version: version,
                  test_config_file_path: testConfigPath,
                  hub_id: testConfig.hub_id || `${publisher}/${connectorName}`,
                  port: testConfig.port || 8083,
                  setup_compose_path: testConfig.setup_compose_path,
                  envs: testConfig.envs || []
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
    
    // Start setup services if needed
    if (connector.setup_compose_path) {
      console.log(`🐳 Starting setup services: ${connector.setup_compose_path}`);
      
      const repoRoot = process.env.NDC_HUB_GIT_REPO_FILE_PATH || process.cwd();
      const testConfigDir = join(repoRoot, connector.test_config_file_path, '..');
      const composePath = join(testConfigDir, connector.setup_compose_path);
      
      await runDockerCommand([
        'compose',
        '-f', composePath,
        '--project-name', `setup-${connector.connector_name}`,
        'up', '-d', '--build', '--wait'
      ]);
      
      // Connect setup services to network
      const composeServices = await new Promise<string[]>((resolve, reject) => {
        const childProcess = spawn('docker', [
          'compose',
          '-f', composePath,
          '--project-name', `setup-${connector.connector_name}`,
          'ps', '--services'
        ], { stdio: 'pipe' });
        
        let output = '';
        childProcess.stdout.on('data', (data) => {
          output += data.toString();
        });
        
        childProcess.on('close', (code) => {
          if (code === 0) {
            resolve(output.trim().split('\n').filter(s => s.length > 0));
          } else {
            reject(new Error(`Failed to get compose services`));
          }
        });
      });
      
      for (const service of composeServices) {
        const serviceContainerName = `setup-${connector.connector_name}-${service}-1`;
        await runDockerCommand(['network', 'connect', networkName, serviceContainerName]).catch(() => {
          // Ignore connection errors
        });
      }
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
    for (const env of connector.envs || []) {
      envVars.push('-e', env);
    }
    
    // Start DDN workspace container
    await runDockerCommand([
      'run', '-d',
      '--name', containerName,
      '--network', networkName,
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
    
    if (connector.setup_compose_path) {
      const repoRoot = process.env.NDC_HUB_GIT_REPO_FILE_PATH || process.cwd();
      const testConfigDir = join(repoRoot, connector.test_config_file_path, '..');
      const composePath = join(testConfigDir, connector.setup_compose_path);
      
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
  
  const repoRoot = process.env.NDC_HUB_GIT_REPO_FILE_PATH || process.cwd();
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
    
    // Read metadata to check if add-to-ddn-workspace is true
    const metadataExists = await readFile(metadataPath, 'utf8').then(() => true).catch(() => false);
    
    if (metadataExists) {
      const metadata = JSON.parse(await readFile(metadataPath, 'utf8'));
      if (!metadata['add-to-ddn-workspace']) {
        throw new Error(`${connectorSpec} does not have add-to-ddn-workspace enabled`);
      }
    }
    
    // Read test config
    const testConfig = JSON.parse(await readFile(testConfigPath, 'utf8'));
    
    const connector: ConnectorTestConfig = {
      connector_name: spec.name,
      connector_version: spec.version,
      test_config_file_path: testConfigPath,
      hub_id: testConfig.hub_id || spec.hubId,
      port: testConfig.port || 8083,
      setup_compose_path: testConfig.setup_compose_path,
      envs: testConfig.envs || []
    };
    
    console.log(`✅ Found connector: ${spec.name}:${spec.version}`);
    return [connector];
    
  } catch (error) {
    console.error(`❌ Error finding connector ${connectorSpec}: ${error}`);
    throw error;
  }
}

export async function runDDNWorkspaceTestSuite(pattern: string = "*", customVersions?: ConnectorVersionOverride[], versionPattern?: string): Promise<void> {
  console.log('🚀 Starting DDN Workspace Test Suite');
  
  let connectors: ConnectorTestConfig[];
  
  // Check if pattern looks like a specific connector spec (contains : and /)
  if (pattern.includes(':') && pattern.includes('/')) {
    connectors = await findSpecificConnector(pattern);
  } else {
    // Find connectors using pattern matching
    connectors = await findConnectorsToTest(pattern, versionPattern);
  }
  
  if (connectors.length === 0) {
    console.log('⚠️  No connectors found matching the pattern');
    return;
  }
  
  console.log(`📋 Found ${connectors.length} connectors to test`);
  
  // Build DDN workspace
  // await buildDDNWorkspace(customVersions);
  
  // Test each connector
  for (const connector of connectors) {
    try {
      await testConnector(connector);
    } catch (error) {
      console.error(`❌ Test failed for ${connector.connector_name}:${connector.connector_version}`, error);
      throw error;
    }
  }
  
  console.log('🎉 All DDN workspace tests completed successfully!');
}
