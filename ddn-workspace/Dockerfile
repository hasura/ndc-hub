######### Arg definitions  #########
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
####################################

######### Builders Begin #########
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
######### Builders End   ##########

FROM --platform=amd64 ubuntu:24.04 AS workspace-builder

# Use CODE_SERVER_VERSION from https://github.com/coder/code-server/releases
ARG CODE_SERVER_VERSION=4.101.2
# Use CLI_VERSION from https://github.com/hasura/v3-cli-go/releases
ARG CLI_VERSION=v3.1.0

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

RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install curl git -y \
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

COPY install-code-server.sh entrypoint.sh ${BIN_PATH}/
RUN chmod +x ${BIN_PATH}/install-code-server.sh \
    && chmod +x ${BIN_PATH}/entrypoint.sh \
    && ${BIN_PATH}/install-code-server.sh "${CODE_SERVER_VERSION}"

######### Install Runtimes #########
RUN curl https://mise.run | MISE_INSTALL_PATH=${BIN_PATH}/mise sh \
    && echo 'eval "$(mise activate bash)"' >> ~/.bashrc \
    && mise use -g node@20  \
    && mise use -g java@openjdk-21

ENV PATH="${XDG_DATA_HOME}/mise/shims:${PATH}"

######### Use latest vulnerability patched npm #########
RUN npm update -g npm

######### Install DDN CLI   #########
RUN curl -fsSL https://graphql-engine-cdn.hasura.io/ddn/cli/v4/get.sh | VERSION=${CLI_VERSION} INSTALL_PATH=${BIN_PATH} bash

######### Nodejs Lambda - NoOP #########

######### Oracle Begin #########
ARG ORACLE_VERSION
COPY --from=oracle-connector  /app ${LIB_PATH}/hasura/oracle/${ORACLE_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}
COPY --from=oracle-cli  /app ${LIB_PATH}/hasura/oracle/${ORACLE_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}
######### Oracle End   #########

######### TRINO Begin #########
ARG TRINO_VERSION
COPY --from=trino-connector  /app ${LIB_PATH}/hasura/trino/${TRINO_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}
COPY --from=trino-cli  /app ${LIB_PATH}/hasura/trino/${TRINO_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}
######### TRINO End   #########

######### Postgres Begin #########
ARG POSTGRES_VERSION
COPY --from=postgres  /bin/ndc-postgres ${LIB_PATH}/hasura/postgres/${POSTGRES_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-postgres
# Required as the postgres binary is hardcoded to look for ld-linux-x86-64.so.2 with nix (interpreter /nix/store/3dyw...)
# Use `ldd ndc-postgres` to check the dynamic dependencies
RUN mkdir -p /nix/store/pacbfvpzqz2mksby36awvbcn051zcji3-glibc-2.40-36/lib \
    && ln -s /lib64/ld-linux-x86-64.so.2 /nix/store/pacbfvpzqz2mksby36awvbcn051zcji3-glibc-2.40-36/lib/ld-linux-x86-64.so.2 \
    && chmod +x ${LIB_PATH}/hasura/postgres/${POSTGRES_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-postgres  \
    && ddn plugin install ndc-postgres --version ${POSTGRES_VERSION} \
    && mkdir -p ${LIB_PATH}/hasura/postgres/${POSTGRES_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && ln -s $HOME/.ddn/plugins/store/ndc-postgres/$POSTGRES_VERSION/hasura-ndc-postgres ${LIB_PATH}/hasura/postgres/${POSTGRES_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/hasura-ndc-postgres
######### Postgres End   #########

######### MongoDB Begin #########
ARG MONGODB_VERSION
COPY --from=mongodb  /nix/store/ /tmp/mongodb-connector/
# Required as the mongo binary is hardcoded to look for ld-linux-x86-64.so.2 with nix (interpreter /nix/store/6q2m...)

RUN mkdir -p ${LIB_PATH}/hasura/mongodb/${MONGODB_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR} \
    && mv /tmp/mongodb-connector/*/bin/mongodb-connector ${LIB_PATH}/hasura/mongodb/${MONGODB_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/mongodb-connector \
    && rm -r /tmp/mongodb-connector \
    && mkdir -p /nix/store/cg9s562sa33k78m63njfn1rw47dp9z0i-glibc-2.40-66/lib \
    && ln -s /lib64/ld-linux-x86-64.so.2 /nix/store/cg9s562sa33k78m63njfn1rw47dp9z0i-glibc-2.40-66/lib/ld-linux-x86-64.so.2 \
    && chmod +x ${LIB_PATH}/hasura/mongodb/${MONGODB_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/mongodb-connector \
    && mkdir -p ${LIB_PATH}/hasura/mongodb/${MONGODB_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && ddn plugin install ndc-mongodb --version ${MONGODB_VERSION} \
    && ln -s $HOME/.ddn/plugins/store/ndc-mongodb/$MONGODB_VERSION/hasura-ndc-mongodb ${LIB_PATH}/hasura/mongodb/${MONGODB_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/hasura-ndc-mongodb
######### MongoDB End   #########

######### GraphQL Begin #########
ARG GRAPHQL_VERSION
COPY --from=graphql /usr/local/bin/ndc-graphql ${LIB_PATH}/hasura/graphql/${GRAPHQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-graphql
RUN ddn plugin install ndc-graphql --version ${GRAPHQL_VERSION} \
    && mkdir -p ${LIB_PATH}/hasura/graphql/${GRAPHQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && ln -s $HOME/.ddn/plugins/store/ndc-graphql/$GRAPHQL_VERSION/hasura-ndc-graphql ${LIB_PATH}/hasura/graphql/${GRAPHQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/hasura-ndc-graphql
######### GraphQL End   #########

######### BigQuery Begin #########
ARG BIGQUERY_VERSION
COPY --from=bigquery  /bin/ndc-bigquery ${LIB_PATH}/hasura/bigquery/${BIGQUERY_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-bigquery
# Required as the bigquery binary is hardcoded to look for ld-linux-x86-64.so.2 with nix (interpreter /nix/store/3dyw...)
# Use -f to force it as Postgres also has the exact symlink
RUN mkdir -p /nix/store/3dyw8dzj9ab4m8hv5dpyx7zii8d0w6fi-glibc-2.39-52/lib \
    && ln -sf /lib64/ld-linux-x86-64.so.2 /nix/store/3dyw8dzj9ab4m8hv5dpyx7zii8d0w6fi-glibc-2.39-52/lib/ld-linux-x86-64.so.2 \
    && chmod +x ${LIB_PATH}/hasura/bigquery/${BIGQUERY_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-bigquery \
    && ddn plugin install ndc-bigquery --version ${BIGQUERY_VERSION} \
    && mkdir -p ${LIB_PATH}/hasura/bigquery/${BIGQUERY_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && ln -s $HOME/.ddn/plugins/store/ndc-bigquery/$BIGQUERY_VERSION/hasura-ndc-bigquery ${LIB_PATH}/hasura/bigquery/${BIGQUERY_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/hasura-ndc-bigquery
######### BigQuery End   #########

######### OpenAPI Begin #########
ARG OPENAPI_VERSION
COPY --from=openapi-plugin  /app ${LIB_PATH}/hasura/openapi/${OPENAPI_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}
# using install-bin-unsafe as it is runs without tests and is faster
RUN cd ${LIB_PATH}/hasura/openapi/${OPENAPI_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && npm install && npm run install-bin-unsafe \
    && ln -s $(npm config get prefix)/bin/ndc-oas-lambda ${LIB_PATH}/hasura/openapi/${OPENAPI_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/ndc-oas-lambda
######### OpenAPI End   #########

######### Elasticsearch Begin #########
ARG ELASTICSEARCH_VERSION
COPY --from=elasticsearch /app/ndc-elasticsearch ${LIB_PATH}/hasura/elasticsearch/${ELASTICSEARCH_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/ndc-elasticsearch
RUN ddn plugin install elasticsearch --version ${ELASTICSEARCH_VERSION} \
    && mkdir -p ${LIB_PATH}/hasura/elasticsearch/${ELASTICSEARCH_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR} \
    && ln -s $HOME/.ddn/plugins/store/elasticsearch/$ELASTICSEARCH_VERSION/hasura-elasticsearch ${LIB_PATH}/hasura/elasticsearch/${ELASTICSEARCH_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/hasura-elasticsearch
######### Elasticsearch End   #########

######### Postgres PromptQL Begin #########
ARG POSTGRES_PROMPTQL_VERSION
COPY --from=postgres-promptql-connector  /app.jar /app.properties* ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/
RUN printf '#!/bin/sh\n\
    root_path=%s/hasura/postgres-promptql/%s/%s\n\
    if [ -f "$root_path/app.properties" ]; then\n\
    DEFAULT_OPTS=$(grep "jvm.opts" "$root_path/app.properties" | cut -d'"'"'='"'"' -f2-)\n\
    else\n\
    DEFAULT_OPTS=""\n\
    fi\n\
    exec java $DEFAULT_OPTS $JAVA_OPTS -jar "$root_path/app.jar" "$@"\n'\
    "${LIB_PATH}" "${POSTGRES_PROMPTQL_VERSION}" "${CONNECTOR_ARTIFACTS_SUB_DIR}" \
    > ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/entrypoint.sh
RUN chmod +x ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/entrypoint.sh

COPY --from=postgres-promptql-cli  /app.jar /app.properties* ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/
RUN printf '#!/bin/sh\n\
    root_path=%s/hasura/postgres-promptql/%s/%s\n\
    if [ -f "$root_path/app.properties" ]; then\n\
    DEFAULT_OPTS=$(grep "jvm.opts" "$root_path/app.properties" | cut -d'"'"'='"'"' -f2-)\n\
    else\n\
    DEFAULT_OPTS=""\n\
    fi\n\
    exec java $DEFAULT_OPTS $JAVA_OPTS -jar "$root_path/app.jar" "$@"\n'\
    "${LIB_PATH}" "${POSTGRES_PROMPTQL_VERSION}" "${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}" \
    > ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/entrypoint.sh
RUN chmod +x ${LIB_PATH}/hasura/postgres-promptql/${POSTGRES_PROMPTQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/entrypoint.sh
######### Postgres PromptQL End   #########

######### Snowflake Begin #########
ARG SNOWFLAKE_JDBC_VERSION
COPY --from=snowflake-connector  /app.jar /app.properties* ${LIB_PATH}/hasura/snowflake-jdbc/${SNOWFLAKE_JDBC_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/
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

COPY --from=snowflake-cli  /app.jar /app.properties* ${LIB_PATH}/hasura/snowflake-jdbc/${SNOWFLAKE_JDBC_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/
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
######### Snowflake End   #########

######### MySQL PromptQL Begin #########
ARG MYSQL_PROMPTQL_VERSION
COPY --from=mysql-promptql-connector  /app.jar /app.properties* ${LIB_PATH}/hasura/mysql-promptql/${MYSQL_PROMPTQL_VERSION}/${CONNECTOR_ARTIFACTS_SUB_DIR}/
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

COPY --from=mysql-promptql-cli  /app.jar /app.properties* ${LIB_PATH}/hasura/mysql-promptql/${MYSQL_PROMPTQL_VERSION}/${CONNECTOR_PLUGIN_ARTIFACTS_SUB_DIR}/
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
######### MySQL PromptQL End   #########

COPY ./sample-projects/ /sample-projects
COPY scripts/show_supported_connector_versions.sh ${BIN_PATH}/show_supported_connector_versions.sh
RUN chmod +x ${BIN_PATH}/show_supported_connector_versions.sh
RUN chgrp -R 0 ${HOME} /workspace /sample-projects \
    && chmod -R g=u ${HOME} /workspace /sample-projects

##########################################################

FROM --platform=amd64 ubuntu:24.04

ENV HOME=/home/hasura
ENV BIN_PATH=${HOME}/.local/bin
ENV PATH="${BIN_PATH}:${PATH}"
ENV LIB_PATH=${HOME}/.local/lib
ENV XDG_CONFIG_HOME=${HOME}/.config
ENV XDG_DATA_HOME=${HOME}/.local/share
ENV XDG_STATE_HOME=${HOME}/.local/state
ENV XDG_CACHE_HOME=${HOME}/.cache

ENV SHELL=/bin/bash

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
    && chsh -s /bin/bash 

COPY --from=workspace-builder ${HOME} ${HOME}
COPY --from=workspace-builder /nix/store /nix/store

ENV CODE_SERVER_PORT=8123
ENV DDN_WORKSPACE=true
ENV DDN_USE_NATIVE_RUNTIME=true
ENV NATIVE_RUNTIME_CONNECTOR_ARTIFACT_ROOT=${LIB_PATH}

EXPOSE $CODE_SERVER_PORT

WORKDIR /workspace

USER 1001

ENTRYPOINT [ "entrypoint.sh" ]
