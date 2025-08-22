# DDN Workspace

A Docker image containing all connectors with DDN workspace support for native runtime testing and development.

## Quick Start

### Pull and Run
```bash
# Pull the latest image (when available in production)
docker pull gcr.io/hasura-ee/ddn-native-workspace:latest

# Run the workspace
docker run -it --privileged gcr.io/hasura-ee/ddn-native-workspace:latest
```

### Local Development
```bash
# Build locally
docker build -t ddn-workspace:local .

# Run with environment variables
docker run -it --privileged \
  -e DATABASE_URL="postgresql://..." \
  -e API_KEY="your-key" \
  ddn-workspace:local
```

## What's Included

This image contains:
- **DDN CLI** with native runtime support
- **All DDN-enabled connectors** from the registry
- **Docker-in-Docker** for connector runtime
- **Testing tools** for connector validation

### Supported Connectors

The image automatically includes all connectors with `ddn_workspace.enabled: true` in their test configuration:

- `hasura/postgres` - PostgreSQL connector
- `hasura/snowflake-jdbc` - Snowflake JDBC connector  
- `hasura/mysql` - MySQL connector
- `hasura/mongodb` - MongoDB connector
- `hasura/elasticsearch` - Elasticsearch connector
- And more...

## Usage

### DDN Workspace Testing
```bash
# Inside the container
ddn connector introspect --connector postgres --config config.json
```

### Environment Variables

Common environment variables you may need:
- `DATABASE_URL` - Database connection string
- `API_KEY` - API authentication key
- `DDN_WORKSPACE_ACCESS_TOKEN` - DDN workspace access token

## Adding Your Connector

To include your connector in future DDN workspace builds:

1. **Enable DDN workspace** in your `tests/test-config.json`:
   ```json
   {
     "ddn_workspace": {
       "enabled": true,
       "envs": ["DATABASE_URL=$YOUR_DB_URL"]
     }
   }
   ```

2. **Publish a release** to the registry

3. **Automatic inclusion** in the next DDN workspace build

See [DDN Workspace Connector Integration Guide](../docs/ddn-workspace-connector-integration.md) for detailed instructions.

## Build Process

### Automated Builds
- Triggered when DDN-enabled connectors are added/updated
- Uses latest versions of all DDN-enabled connectors
- Builds with `scripts/build-with-versions.sh`

### Manual Build
```bash
# Build with specific connector versions
./scripts/build-with-versions.sh connector-versions.json ddn-workspace:test production
```

## Architecture

```
DDN Workspace Container
├── DDN CLI (native runtime)
├── Docker daemon (for connector runtime)
├── Connector binaries
│   ├── hasura/postgres:v3.1.0
│   ├── hasura/snowflake-jdbc:v1.2.12
│   └── ...
└── Testing tools
```

## Development

### Local Testing
```bash
# Test a specific connector
cd registry-automation/e2e-testing
export DDN_WORKSPACE_ACCESS_TOKEN='your-token'
bun ddn-workspace-testing.ts "hasura/postgres:v3.1.0"
```

### Build Scripts
- `scripts/build-with-versions.sh` - Build with specific connector versions
- `scripts/test-connectors.sh` - Test all included connectors

## Troubleshooting

### Common Issues

**Container won't start:**
- Ensure `--privileged` flag for Docker-in-Docker
- Check available disk space

**Connector not found:**
- Verify connector is DDN-enabled in test config
- Check if connector version exists in releases

**Environment variables not working:**
- Use `-e VAR=value` format when running
- Check variable names match connector requirements

### Logs
```bash
# View container logs
docker logs <container-id>

# Debug inside container
docker exec -it <container-id> bash
```

## Contributing

1. Add/update connector configurations
2. Test with DDN workspace testing script
3. Submit PR with changes
4. Automated builds will include your connector

## Support

- Review connector integration guide
- Check GitHub Actions logs for build issues
- Contact DDN team for support
