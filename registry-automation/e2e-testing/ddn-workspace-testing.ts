#!/usr/bin/env bun

import { runDDNWorkspaceTestSuite, ConnectorVersionOverride } from './ddn-workspace';

async function main() { 
  const args = process.argv.slice(2);
   
  if (args.includes('--help') || args.includes('-h')) {
    console.log(`
Usage: bun ddn-workspace-testing.ts [PATTERN] [CUSTOM_VERSIONS]

Examples:
  bun ddn-workspace-testing.ts                         
  bun ddn-workspace-testing.ts "hasura/elasticsearch:v1.9.5"  # Run specific connector version

Arguments:
  PATTERN         Glob pattern to match connector directories (default: "*")
  CUSTOM_VERSIONS JSON array of connector version overrides

Environment Variables:
  HASURA_DDN_PAT                 Required: Hasura DDN access token
  DDN_WORKSPACE_ACCESS_TOKEN     Optional: Alternative to HASURA_DDN_PAT
`);
    process.exit(0);
  }

  // Check required environment variables
  if (!process.env.HASURA_DDN_PAT && !process.env.DDN_WORKSPACE_ACCESS_TOKEN) {
    console.error('❌ Error: HASURA_DDN_PAT or DDN_WORKSPACE_ACCESS_TOKEN environment variable is required');
    process.exit(1);
  }

  const pattern = args[0] || "*";
  let customVersions: ConnectorVersionOverride[] | undefined;

  if (args[1]) {
    try {
      customVersions = JSON.parse(args[1]);
      console.log('📝 Using custom connector versions:', customVersions);
    } catch (error) {
      console.error('❌ Error: Invalid JSON for custom versions:', error);
      process.exit(1);
    }
  }

  try {
    console.log(`🚀 Starting DDN workspace test suite with pattern: ${pattern}`);
    await runDDNWorkspaceTestSuite(pattern, customVersions);
    console.log('✅ DDN workspace test suite completed successfully!');
  } catch (error) {
    console.error('❌ DDN workspace test suite failed:', error);
    process.exit(1);
  }
}

main().catch(console.error);
