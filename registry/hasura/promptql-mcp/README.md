# PromptQL MCP Connector

The PromptQL MCP Connector bridges PromptQL with Model Context Protocol (MCP) servers, enabling you to expose MCP resources as collections and MCP tools as functions/procedures.
This connector provides seamless integration with AI-powered tools and data sources through the standardized MCP protocol.

## Features

- **Dynamic Schema Generation**: Automatically generates NDC schema from MCP server introspection
- **Multiple Transport Support**: Connect via stdio (local processes) and HTTP (remote servers)
- **Multi-Server Support**: Connect to multiple MCP servers simultaneously
- **Resource Mapping**: MCP resources are exposed as NDC collections
- **Tool Execution**: MCP tools are exposed as NDC functions (read-only) and procedures (mutable)

## Prerequisites

1. Create a [Hasura Cloud account](https://console.hasura.io)
2. Please ensure you have the [DDN CLI](https://hasura.io/docs/3.0/cli/installation) and
   [Docker](https://docs.docker.com/engine/install/) installed
3. [Create a supergraph](https://hasura.io/docs/3.0/getting-started/init-supergraph)
4. [Create a subgraph](https://hasura.io/docs/3.0/getting-started/init-subgraph)
5. One or more MCP servers to connect to

The steps below explain how to initialize and configure a connector on your local machine (typically for development
purposes). You can learn how to deploy a connector to Hasura DDN — after it's been configured —
[here](https://hasura.io/docs/3.0/getting-started/deployment/deploy-a-connector).

## Using the connector

### Initialize the connector

```bash
ddn connector init -i
```

Name the connector `promptql_mcp`.
Choose the hasura/promptql-mcp connector from the list.

### Configuration

The connector uses a JSON configuration file to define MCP server connections. Create a `configuration.json` file in your connector directory:

#### Stdio Transport (Local Processes)

```json
{
  "servers": {
    "filesystem": {
      "type": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "/path/to/directory"
      ],
      "env": {
        "DEBUG": "true",
        "API_KEY": {
          "fromEnv": "FILESYSTEM_API_KEY"
        }
      },
      "env_file": ".env"
    },
    "sequential_thinking": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"],
      "env": {
        "LOG_LEVEL": {
          "fromEnv": "MCP_LOG_LEVEL"
        }
      }
    }
  }
}
```

#### HTTP Transport (Remote Servers)

```json
{
  "servers": {
    "remote_server": {
      "type": "http",
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": {
          "fromEnv": "MCP_AUTH_TOKEN"
        },
        "Content-Type": "application/json",
        "X-API-Key": {
          "fromEnv": "MCP_API_KEY"
        }
      },
      "timeout_seconds": 30
    }
  }
}
```

#### Mixed Configuration

```json
{
  "servers": {
    "local_filesystem": {
      "type": "stdio",
      "command": "mcp-filesystem-server",
      "args": ["/home/user/documents"],
      "env": {
        "FILESYSTEM_ROOT": {
          "fromEnv": "FILESYSTEM_ROOT_PATH"
        }
      }
    },
    "remote_api": {
      "type": "http",
      "url": "https://api.example.com/mcp",
      "headers": {
        "Authorization": {
          "fromEnv": "REMOTE_API_TOKEN"
        }
      },
      "timeout_seconds": 60
    }
  }
}
```

### Configuration Parameters

#### Stdio Transport

- `type`: Must be "stdio"
- `command`: Executable command to run the MCP server
- `args`: Array of command-line arguments
- `env`: Environment variables (optional) - supports both literal values and `fromEnv` references
- `env_file`: Path to a .env file to load additional environment variables (optional)

#### HTTP Transport

- `type`: Must be "http"
- `url`: HTTP endpoint URL for the MCP server
- `headers`: HTTP headers to include in requests (optional) - supports both literal values and `fromEnv` references
- `timeout_seconds`: Request timeout in seconds (optional, default: 30)

#### Environment Variable Configuration

The connector supports two ways to specify environment variables and headers:

**Literal Values:**

```json
{
  "env": {
    "DEBUG": "true",
    "LOG_LEVEL": "info"
  }
}
```

**Environment Variable References:**

```json
{
  "env": {
    "API_KEY": {
      "fromEnv": "MY_API_KEY"
    },
    "DATABASE_URL": {
      "fromEnv": "DB_CONNECTION_STRING"
    }
  }
}
```

The `fromEnv` syntax allows you to reference environment variables from the DDN .env files, which is useful for:

- Keeping sensitive information like API keys out of configuration files
- Supporting different environments (development, staging, production)
- Following security best practices

**Note:** When using `fromEnv`, the referenced environment variable must be available in the container environment where the connector runs. Use `envMapping` in `connector.yaml` to map DDN .env variables to container environment variables.

### Installing MCP Servers

If you need to install MCP servers that aren't available in the base image, you can extend the Dockerfile. The connector includes a comprehensive Dockerfile template with examples for different runtime environments.

#### Extending the Dockerfile

The base Dockerfile (`connector-definition/.hasura-connector/Dockerfile`) provides extension points for installing additional MCP servers:

**For Node.js-based MCP servers:**

```dockerfile
# Uncomment these lines in your Dockerfile
USER root
RUN apt-get update && apt-get install -y \
    nodejs \
    npm \
    && rm -rf /var/lib/apt/lists/*
USER connector

# Install specific MCP servers
RUN npm install -g @modelcontextprotocol/server-filesystem
RUN npm install -g @modelcontextprotocol/server-sequential-thinking
RUN npm install -g @modelcontextprotocol/server-brave-search
```

**For Python-based MCP servers:**

```dockerfile
# Uncomment these lines in your Dockerfile
USER root
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    git \
    && rm -rf /var/lib/apt/lists/*
USER connector

# Create Python virtual environment (recommended)
RUN python3 -m venv /home/connector/.venv
ENV PATH="/home/connector/.venv/bin:$PATH"

# Install Python MCP servers
RUN pip install mcp-server-git
RUN pip install mcp-server-sqlite
```

**Custom MCP servers:**

```dockerfile
# Install custom dependencies
USER root
RUN apt-get update && apt-get install -y \
    your-custom-dependencies \
    && rm -rf /var/lib/apt/lists/*
USER connector

# Copy and install your custom MCP server
COPY --chown=connector:connector ./my-custom-mcp-server /opt/my-custom-mcp-server
RUN cd /opt/my-custom-mcp-server && npm install
```

#### Important Notes:

- Always switch to `root` user for system package installations
- Switch back to `connector` user for application-level installations
- The base image provides a non-root `connector` user (UID 1000) for security
- Use virtual environments for Python installations to avoid conflicts
- Ensure installed MCP servers are available in the system PATH

### Introspect the connector

```bash
ddn connector introspect promptql_mcp
```

This will discover all available resources and tools from your configured MCP servers and generate the appropriate schema.

### Add resources and tools

```bash
# Add all collections (resources)
ddn model add promptql_mcp "*"

# Add all functions and procedures (tools)
ddn command add promptql_mcp "*"
```

### Build and run locally

```bash
ddn supergraph build local
ddn run docker-start
```

Then visit the console to explore and query your MCP resources through the GraphQL API.

## How it Works

### Resource Mapping

MCP resources are mapped to NDC collections with the naming pattern `{server_name}__{resource_id}`. Each resource becomes queryable as a collection in your schema.

### Tool Mapping

MCP tools are mapped to NDC functions or procedures based on their characteristics:

- **Read-only tools**: Exposed as functions
- **Mutable tools**: Exposed as procedures

Tool names follow the pattern `{server_name}__{tool_id}`.

## Troubleshooting

### Connection Issues

- Verify MCP server binaries are accessible to the connector. Use `docker exec` to check inside the container.
- Check network connectivity for HTTP-based servers
- Validate authentication credentials and headers

### Schema Issues

- Ensure MCP servers support introspection
- Check connector logs for capability negotiation errors
- Verify resource and tool definitions are valid

### Performance

- Consider timeout settings for slow MCP servers
- Monitor resource usage for stdio-based servers
- Use connection pooling for high-traffic scenarios

## Support

Please [submit a GitHub issue](https://github.com/hasura/ndc-mcp-rs/issues/new) if you encounter any problems or have feature requests!

For more information about the Model Context Protocol, visit the [official MCP documentation](https://modelcontextprotocol.io/).
