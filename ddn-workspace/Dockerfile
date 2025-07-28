######### Connector Version Arguments #########
# Default versions - can be overridden with build args
ARG POSTGRES_VERSION=v2.1.1
ARG ORACLE_VERSION=v1.1.1
ARG ORACLE_CLI_VERSION=v1.1.1-oracle
ARG TRINO_VERSION=v1.0.4
ARG TRINO_CLI_VERSION=v1.0.4-trino
ARG MONGODB_VERSION=v1.8.3
ARG GRAPHQL_VERSION=v0.2.3
ARG BIGQUERY_VERSION=v2.0.1
ARG OPENAPI_VERSION=v1.7.1
ARG ELASTICSEARCH_VERSION=v1.9.5
ARG POSTGRES_PROMPTQL_VERSION=v1.2.7
ARG POSTGRES_PROMPTQL_CLI_VERSION=v1.2.7
ARG SNOWFLAKE_JDBC_VERSION=v1.2.12
ARG SNOWFLAKE_JDBC_CLI_VERSION=v1.2.12
ARG MYSQL_PROMPTQL_VERSION=v1.2.7
ARG MYSQL_PROMPTQL_CLI_VERSION=v1.2.7

# Infrastructure versions
ARG CODE_SERVER_VERSION=4.101.2
ARG CLI_VERSION=v3.1.0

######### Connector Image Builders #########
FROM --platform=amd64 ghcr.io/hasura/ndc-postgres:${POSTGRES_VERSION} AS postgres
FROM --platform=amd64 ghcr.io/hasura/ndc-jvm-oracle:${ORACLE_VERSION} AS oracle-connector
FROM --platform=amd64 ghcr.io/hasura/ndc-jvm-cli:${ORACLE_CLI_VERSION} AS oracle-cli
FROM --platform=amd64 ghcr.io/hasura/ndc-jvm-trino:${TRINO_VERSION} AS trino-connector
FROM --platform=amd64 ghcr.io/hasura/ndc-jvm-cli:${TRINO_CLI_VERSION} AS trino-cli
FROM --platform=amd64 ghcr.io/hasura/ndc-mongodb:${MONGODB_VERSION} AS mongodb
FROM --platform=amd64 ghcr.io/hasura/ndc-graphql:${GRAPHQL_VERSION} AS graphql
FROM --platform=amd64 ghcr.io/hasura/ndc-bigquery:${BIGQUERY_VERSION} AS bigquery
FROM --platform=amd64 ghcr.io/hasura/ndc-open-api-lambda:${OPENAPI_VERSION} AS openapi-plugin
FROM --platform=amd64 ghcr.io/hasura/ndc-elasticsearch:${ELASTICSEARCH_VERSION} AS elasticsearch
FROM --platform=amd64 ghcr.io/hasura/ndc-postgres-jdbc:${POSTGRES_PROMPTQL_VERSION} AS postgres-promptql-connector
FROM --platform=amd64 ghcr.io/hasura/ndc-postgres-jdbc-cli:${POSTGRES_PROMPTQL_CLI_VERSION} AS postgres-promptql-cli
FROM --platform=amd64 ghcr.io/hasura/ndc-snowflake-jdbc:${SNOWFLAKE_JDBC_VERSION} AS snowflake-connector
FROM --platform=amd64 ghcr.io/hasura/ndc-snowflake-jdbc-cli:${SNOWFLAKE_JDBC_CLI_VERSION} AS snowflake-cli
FROM --platform=amd64 ghcr.io/hasura/ndc-mysql-jdbc:${MYSQL_PROMPTQL_VERSION} AS mysql-promptql-connector
FROM --platform=amd64 ghcr.io/hasura/ndc-mysql-jdbc-cli:${MYSQL_PROMPTQL_CLI_VERSION} AS mysql-promptql-cli

######### DDN Registry Connector Builders #########
# Build postgres-promptql using registry Dockerfile.ddn if it exists
FROM --platform=amd64 ubuntu:24.04 AS postgres-promptql-ddn-base

# Set up basic environment for connector building
ENV HOME=/home/hasura
ENV BIN_PATH=${HOME}/.local/bin
ENV LIB_PATH=${HOME}/.local/lib
ENV PATH="${BIN_PATH}:${PATH}"

RUN apt-get update && apt-get install -y curl git jq && \
    mkdir -p $HOME $BIN_PATH $LIB_PATH && \
    useradd -m -s /bin/bash -u 1001 hasura

# Check if postgres-promptql has a Dockerfile.ddn and build it
FROM postgres-promptql-ddn-base AS postgres-promptql-ddn
ARG POSTGRES_PROMPTQL_VERSION

# Copy registry files to check for Dockerfile.ddn
COPY registry/ /tmp/registry/

# Build postgres-promptql using Dockerfile.ddn if it exists, otherwise skip
RUN if [ -f "/tmp/registry/hasura/postgres-promptql/Dockerfile.ddn" ]; then \
        echo "Found Dockerfile.ddn for postgres-promptql, building..."; \
        # Get source images from packaging file \
        PACKAGING_FILE="/tmp/registry/hasura/postgres-promptql/releases/${POSTGRES_PROMPTQL_VERSION}/connector-packaging.json"; \
        if [ -f "$PACKAGING_FILE" ]; then \
            SOURCE_IMAGE=$(jq -r '.packagingDefinition.dockerImage' "$PACKAGING_FILE"); \
            CLI_IMAGE=$(echo "$SOURCE_IMAGE" | sed 's/:/-cli:/'); \
            echo "Source images: $SOURCE_IMAGE, $CLI_IMAGE"; \
            # Here we would build the connector, but since we can't do nested docker builds, \
            # we'll mark this for external building \
            echo "$SOURCE_IMAGE" > /tmp/postgres-promptql-source.txt; \
            echo "$CLI_IMAGE" > /tmp/postgres-promptql-cli-source.txt; \
        fi; \
    else \
        echo "No Dockerfile.ddn found for postgres-promptql, will use legacy build"; \
        touch /tmp/use-legacy-postgres-promptql; \
    fi

######### Main Workspace Image #########
FROM --platform=amd64 ubuntu:24.04

# Environment setup
ENV HOME=/home/hasura
ENV BIN_PATH=${HOME}/.local/bin
ENV PATH="${BIN_PATH}:${PATH}"
ENV LIB_PATH=${HOME}/.local/lib
ENV XDG_CONFIG_HOME=${HOME}/.config
ENV XDG_DATA_HOME=${HOME}/.local/share
ENV XDG_STATE_HOME=${HOME}/.local/state
ENV XDG_CACHE_HOME=${HOME}/.cache
ENV CONNECTOR_ARTIFACTS_SUB_DIR="connector"
ENV CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR="plugin"
ENV SHELL=/bin/bash

# System dependencies and directory setup
RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install curl git jq openssl ca-certificates -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p $HOME \
    && mkdir -p /workspace \
    && mkdir -p /sample-projects \
    && mkdir -p $BIN_PATH  \
    && mkdir -p $LIB_PATH  \
    && mkdir -p $XDG_CONFIG_HOME \
    && mkdir -p $XDG_DATA_HOME  \
    && mkdir -p $XDG_STATE_HOME \
    && mkdir -p $XDG_CACHE_HOME \
    && chsh -s /bin/bash \
    && chgrp -R 0 ${HOME} /workspace /sample-projects && chmod -R g=u ${HOME} /workspace /sample-projects \
    && curl -fL  https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -o /usr/bin/yq && \
    chmod +x /usr/bin/yq

# Install VS Code Server
COPY scripts/install-code-server.sh ${BIN_PATH}/
ARG CODE_SERVER_VERSION
RUN chmod +x ${BIN_PATH}/install-code-server.sh \
    && ${BIN_PATH}/install-code-server.sh "${CODE_SERVER_VERSION}"

# Install runtime environments (Node.js, Java)
RUN curl https://mise.run | MISE_INSTALL_PATH=${BIN_PATH}/mise sh \
    && echo 'eval "$(mise activate bash)"' >> ~/.bashrc \
    && mise use -g node@20  \
    && mise use -g java@openjdk-21
ENV PATH="${XDG_DATA_HOME}/mise/shims:${PATH}"

# Update npm to latest version
RUN npm update -g npm

# Install DDN CLI
ARG CLI_VERSION
RUN curl -fsSL https://graphql-engine-cdn.hasura.io/ddn/cli/v4/get.sh | VERSION=${CLI_VERSION} INSTALL_PATH=${BIN_PATH} bash

######### Install Connectors #########

# Copy registry to check for Dockerfile.ddn files
COPY registry/ /tmp/registry/

# Helper script to build connectors with DDN support
COPY scripts/build-connector-with-ddn.sh ${BIN_PATH}/
RUN chmod +x ${BIN_PATH}/build-connector-with-ddn.sh

# Postgres (legacy build - no Dockerfile.ddn yet)
ARG POSTGRES_VERSION
COPY --from=postgres /bin/ndc-postgres ${LIB_PATH}/hasura/postgres/${POSTGRES_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-postgres
RUN mkdir -p /nix/store/pacbfvpzqz2mksby36awvbcn051zcji3-glibc-2.40-36/lib \
    && ln -s /lib64/ld-linux-x86-64.so.2 /nix/store/pacbfvpzqz2mksby36awvbcn051zcji3-glibc-2.40-36/lib/ld-linux-x86-64.so.2 \
    && chmod +x ${LIB_PATH}/hasura/postgres/${POSTGRES_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-postgres  \
    && ddn plugin install ndc-postgres --version ${POSTGRES_VERSION} \
    && mkdir -p ${LIB_PATH}/hasura/postgres/${POSTGRES_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && ln -s $HOME/.ddn/plugins/store/ndc-postgres/$POSTGRES_VERSION/hasura-ndc-postgres ${LIB_PATH}/hasura/postgres/${POSTGRES_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/hasura-ndc-postgres

# Oracle (legacy build)
ARG ORACLE_VERSION
COPY --from=oracle-connector /app ${LIB_PATH}/hasura/oracle/${ORACLE_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}
COPY --from=oracle-cli /app ${LIB_PATH}/hasura/oracle/${ORACLE_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}

# Trino (legacy build)
ARG TRINO_VERSION
COPY --from=trino-connector /app ${LIB_PATH}/hasura/trino/${TRINO_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}
COPY --from=trino-cli /app ${LIB_PATH}/hasura/trino/${TRINO_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}

# MongoDB (legacy build)
ARG MONGODB_VERSION
COPY --from=mongodb /nix/store/ /tmp/mongodb-connector/
RUN mkdir -p ${LIB_PATH}/hasura/mongodb/${MONGODB_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR} \
    && mv /tmp/mongodb-connector/*/bin/mongodb-connector ${LIB_PATH}/hasura/mongodb/${MONGODB_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/mongodb-connector \
    && rm -r /tmp/mongodb-connector \
    && mkdir -p /nix/store/cg9s562sa33k78m63njfn1rw47dp9z0i-glibc-2.40-66/lib \
    && ln -s /lib64/ld-linux-x86-64.so.2 /nix/store/cg9s562sa33k78m63njfn1rw47dp9z0i-glibc-2.40-66/lib/ld-linux-x86-64.so.2 \
    && chmod +x ${LIB_PATH}/hasura/mongodb/${MONGODB_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/mongodb-connector \
    && mkdir -p ${LIB_PATH}/hasura/mongodb/${MONGODB_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && ddn plugin install ndc-mongodb --version ${MONGODB_VERSION} \
    && ln -s $HOME/.ddn/plugins/store/ndc-mongodb/$MONGODB_VERSION/hasura-ndc-mongodb ${LIB_PATH}/hasura/mongodb/${MONGODB_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/hasura-ndc-mongodb

# GraphQL (legacy build)
ARG GRAPHQL_VERSION
COPY --from=graphql /usr/local/bin/ndc-graphql ${LIB_PATH}/hasura/graphql/${GRAPHQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-graphql
RUN ddn plugin install ndc-graphql --version ${GRAPHQL_VERSION} \
    && mkdir -p ${LIB_PATH}/hasura/graphql/${GRAPHQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && ln -s $HOME/.ddn/plugins/store/ndc-graphql/$GRAPHQL_VERSION/hasura-ndc-graphql ${LIB_PATH}/hasura/graphql/${GRAPHQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/hasura-ndc-graphql

# BigQuery (legacy build)
ARG BIGQUERY_VERSION
COPY --from=bigquery /bin/ndc-bigquery ${LIB_PATH}/hasura/bigquery/${BIGQUERY_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-bigquery
RUN mkdir -p /nix/store/3dyw8dzj9ab4m8hv5dpyx7zii8d0w6fi-glibc-2.39-52/lib \
    && ln -sf /lib64/ld-linux-x86-64.so.2 /nix/store/3dyw8dzj9ab4m8hv5dpyx7zii8d0w6fi-glibc-2.39-52/lib/ld-linux-x86-64.so.2 \
    && chmod +x ${LIB_PATH}/hasura/bigquery/${BIGQUERY_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-bigquery \
    && ddn plugin install ndc-bigquery --version ${BIGQUERY_VERSION} \
    && mkdir -p ${LIB_PATH}/hasura/bigquery/${BIGQUERY_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && ln -s $HOME/.ddn/plugins/store/ndc-bigquery/$BIGQUERY_VERSION/hasura-ndc-bigquery ${LIB_PATH}/hasura/bigquery/${BIGQUERY_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/hasura-ndc-bigquery

# OpenAPI (legacy build)
ARG OPENAPI_VERSION
COPY --from=openapi-plugin /app ${LIB_PATH}/hasura/openapi/${OPENAPI_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}
RUN cd ${LIB_PATH}/hasura/openapi/${OPENAPI_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && npm install && npm run install-bin-unsafe \
    && ln -s $(npm config get prefix)/bin/ndc-oas-lambda ${LIB_PATH}/hasura/openapi/${OPENAPI_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/ndc-oas-lambda

# Elasticsearch (legacy build)
ARG ELASTICSEARCH_VERSION
COPY --from=elasticsearch /app/ndc-elasticsearch ${LIB_PATH}/hasura/elasticsearch/${ELASTICSEARCH_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-elasticsearch
RUN ddn plugin install elasticsearch --version ${ELASTICSEARCH_VERSION} \
    && mkdir -p ${LIB_PATH}/hasura/elasticsearch/${ELASTICSEARCH_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && ln -s $HOME/.ddn/plugins/store/elasticsearch/$ELASTICSEARCH_VERSION/hasura-elasticsearch ${LIB_PATH}/hasura/elasticsearch/${ELASTICSEARCH_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/hasura-elasticsearch

# Postgres PromptQL - Check for Dockerfile.ddn first
ARG POSTGRES_PROMPTQL_VERSION
RUN ${BIN_PATH}/build-connector-with-ddn.sh postgres-promptql ${POSTGRES_PROMPTQL_VERSION}

# Fallback to legacy postgres-promptql build if no Dockerfile.ddn
COPY --from=postgres-promptql-connector /app.jar /app.properties* /tmp/postgres-promptql-legacy-connector/
COPY --from=postgres-promptql-cli /app.jar /app.properties* /tmp/postgres-promptql-legacy-cli/
RUN if [ ! -d "${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}" ]; then \
        echo "Using legacy postgres-promptql build"; \
        mkdir -p ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}; \
        mkdir -p ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}; \
        cp /tmp/postgres-promptql-legacy-connector/* ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/; \
        cp /tmp/postgres-promptql-legacy-cli/* ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/; \
        printf '#!/bin/sh\nroot_path=%s/hasura/postgres-promptql/%s/%s\nif [ -f "$root_path/app.properties" ]; then\nDEFAULT_OPTS=$(grep "jvm.opts" "$root_path/app.properties" | cut -d'"'"'='"'"' -f2-)\nelse\nDEFAULT_OPTS=""\nfi\nexec java $DEFAULT_OPTS $JAVA_OPTS -jar "$root_path/app.jar" "$@"\n' \
            "${LIB_PATH}" "${POSTGRES_PROMPTQL_VERSION}" "${CONNECTOR_ARTIFACTS_SUB_DIR}" \
            > ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/entrypoint.sh; \
        chmod +x ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/entrypoint.sh; \
        printf '#!/bin/sh\nroot_path=%s/hasura/postgres-promptql/%s/%s\nif [ -f "$root_path/app.properties" ]; then\nDEFAULT_OPTS=$(grep "jvm.opts" "$root_path/app.properties" | cut -d'"'"'='"'"' -f2-)\nelse\nDEFAULT_OPTS=""\nfi\nexec java $DEFAULT_OPTS $JAVA_OPTS -jar "$root_path/app.jar" "$@"\n' \
            "${LIB_PATH}" "${POSTGRES_PROMPTQL_VERSION}" "${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}" \
            > ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/entrypoint.sh; \
        chmod +x ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/entrypoint.sh; \
    fi

# Snowflake JDBC (legacy build)
ARG SNOWFLAKE_JDBC_VERSION
COPY --from=snowflake-connector /app.jar /app.properties* ${LIB_PATH}/hasura/snowflake-jdbc/${SNOWFLAKE_JDBC_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/
RUN printf '#!/bin/sh\n\
    root_path=%s/hasura/snowflake-jdbc/%s/%s\n\
    if [ -f "$root_path/app.properties" ]; then\n\
    DEFAULT_OPTS=$(grep "jvm.opts" "$root_path/app.properties" | cut -d'"'"'='"'"' -f2-)\n\
    else\n\
    DEFAULT_OPTS=""\n\
    fi\n\
    exec java $DEFAULT_OPTS $JAVA_OPTS -jar "$root_path/app.jar" "$@"\n'\
    "${LIB_PATH}" "${SNOWFLAKE_JDBC_VERSION}" "${CONNECTOR_ARTIFACTS_SUB_DIR}" \
    > ${LIB_PATH}/hasura/snowflake-jdbc/${SNOWFLAKE_JDBC_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/entrypoint.sh
RUN chmod +x ${LIB_PATH}/hasura/snowflake-jdbc/${SNOWFLAKE_JDBC_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/entrypoint.sh

COPY --from=snowflake-cli /app.jar /app.properties* ${LIB_PATH}/hasura/snowflake-jdbc/${SNOWFLAKE_JDBC_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/
RUN printf '#!/bin/sh\n\
    root_path=%s/hasura/snowflake-jdbc/%s/%s\n\
    if [ -f "$root_path/app.properties" ]; then\n\
    DEFAULT_OPTS=$(grep "jvm.opts" "$root_path/app.properties" | cut -d'"'"'='"'"' -f2-)\n\
    else\n\
    DEFAULT_OPTS=""\n\
    fi\n\
    exec java $DEFAULT_OPTS $JAVA_OPTS -jar "$root_path/app.jar" "$@"\n'\
    "${LIB_PATH}" "${SNOWFLAKE_JDBC_VERSION}" "${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}" \
    > ${LIB_PATH}/hasura/snowflake-jdbc/${SNOWFLAKE_JDBC_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/entrypoint.sh
RUN chmod +x ${LIB_PATH}/hasura/snowflake-jdbc/${SNOWFLAKE_JDBC_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/entrypoint.sh

# MySQL PromptQL (legacy build)
ARG MYSQL_PROMPTQL_VERSION
COPY --from=mysql-promptql-connector /app.jar /app.properties* ${LIB_PATH}/hasura/mysql-promptql/${MYSQL_PROMPTQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/
RUN printf '#!/bin/sh\n\
    root_path=%s/hasura/mysql-promptql/%s/%s\n\
    if [ -f "$root_path/app.properties" ]; then\n\
    DEFAULT_OPTS=$(grep "jvm.opts" "$root_path/app.properties" | cut -d'"'"'='"'"' -f2-)\n\
    else\n\
    DEFAULT_OPTS=""\n\
    fi\n\
    exec java $DEFAULT_OPTS $JAVA_OPTS -jar "$root_path/app.jar" "$@"\n'\
    "${LIB_PATH}" "${MYSQL_PROMPTQL_VERSION}" "${CONNECTOR_ARTIFACTS_SUB_DIR}" \
    > ${LIB_PATH}/hasura/mysql-promptql/${MYSQL_PROMPTQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/entrypoint.sh
RUN chmod +x ${LIB_PATH}/hasura/mysql-promptql/${MYSQL_PROMPTQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/entrypoint.sh

COPY --from=mysql-promptql-cli /app.jar /app.properties* ${LIB_PATH}/hasura/mysql-promptql/${MYSQL_PROMPTQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/
RUN printf '#!/bin/sh\n\
    root_path=%s/hasura/mysql-promptql/%s/%s\n\
    if [ -f "$root_path/app.properties" ]; then\n\
    DEFAULT_OPTS=$(grep "jvm.opts" "$root_path/app.properties" | cut -d'"'"'='"'"' -f2-)\n\
    else\n\
    DEFAULT_OPTS=""\n\
    fi\n\
    exec java $DEFAULT_OPTS $JAVA_OPTS -jar "$root_path/app.jar" "$@"\n'\
    "${LIB_PATH}" "${MYSQL_PROMPTQL_VERSION}" "${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}" \
    > ${LIB_PATH}/hasura/mysql-promptql/${MYSQL_PROMPTQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/entrypoint.sh
RUN chmod +x ${LIB_PATH}/hasura/mysql-promptql/${MYSQL_PROMPTQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/entrypoint.sh

# Clean up temporary files
RUN rm -rf /tmp/registry /tmp/postgres-promptql-legacy-*

# Add connector version discovery script
RUN cat <<'EOF' > ${BIN_PATH}/show_supported_connector_versions.sh
#!/usr/bin/env bash
set -euo pipefail

# Script to list all directories in {LIB_PATH} with their namespace, name, and version
# and output as JSON in the format {"namespace/name": "version", ...}

TEMP_FILE=$(mktemp)
echo "{}" > "$TEMP_FILE"

# Find all namespace/name/version directories
find "$LIB_PATH" -mindepth 3 -maxdepth 3 -type d | while read -r dir; do
  # Extract the relative path from LIB_PATH
  rel_path=${dir#"$LIB_PATH/"}
  
  # Extract namespace, name, and version
  namespace=$(echo "$rel_path" | cut -d '/' -f 1)
  name=$(echo "$rel_path" | cut -d '/' -f 2)
  version=$(echo "$rel_path" | cut -d '/' -f 3)
  
  # Check if version follows semver pattern (either v1.2.3 or 1.2.3)
  if [[ "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    # Create the key in the format "namespace/name"
    key="$namespace/$name"
    
    # Update the JSON object
    cat "$TEMP_FILE" | jq --arg key "$key" --arg val "$version" '. + {($key): $val}' > "$TEMP_FILE.new"
    mv "$TEMP_FILE.new" "$TEMP_FILE"
  fi
done

# Format and output the JSON
jq '.' "$TEMP_FILE"

# Clean up
rm "$TEMP_FILE"
EOF

RUN chmod +x ${BIN_PATH}/show_supported_connector_versions.sh

# Create entrypoint script
RUN cat <<'EOF' > ${BIN_PATH}/entrypoint.sh
#!/usr/bin/env bash
set -euo pipefail

# Check if password auth is requested
if [ "${HASURA_DDN_PAT:-}" ]; then
    echo "Using PAT login for VS Code Server"
    exec code-server --bind-addr 0.0.0.0:${CODE_SERVER_PORT:-8123} --auth none /workspace
elif [ "${PASSWORD:-}" ] && [ "${PASSWORD}" != "false" ]; then
    echo "Starting VS Code Server with password authentication on port ${CODE_SERVER_PORT:-8123}"
    exec code-server --bind-addr 0.0.0.0:${CODE_SERVER_PORT:-8123} --auth password /workspace
else
    echo "Starting VS Code Server without authentication on port ${CODE_SERVER_PORT:-8123}"
    exec code-server --bind-addr 0.0.0.0:${CODE_SERVER_PORT:-8123} --auth none /workspace
fi
EOF

RUN chmod +x ${BIN_PATH}/entrypoint.sh

# Final setup
RUN chgrp -R 0 ${HOME} /workspace /sample-projects \
    && chmod -R g=u ${HOME} /workspace /sample-projects

# Environment variables for runtime
ENV CODE_SERVER_PORT=8123
ENV DDN_WORKSPACE=true
ENV DDN_USE_NATIVE_RUNTIME=true
ENV NATIVE_RUNTIME_CONNECTOR_ARTIFACT_ROOT=${LIB_PATH}

EXPOSE $CODE_SERVER_PORT
WORKDIR /workspace
USER 1001

ENTRYPOINT [ "entrypoint.sh" ]