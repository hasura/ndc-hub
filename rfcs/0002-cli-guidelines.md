### CLI Plugin Guidelines

Connectors can optionally provide a CLI plugin to help with author the
connector's configuration. These are the broad categories of commands that
could be useful:

1. init (optional)

   Initializing the config directory for a connector. There could be
   more than one template.

   Let's ignore this for now (out of beta's scope) as the cli's `add
   ConnectorManifest` would also initialize the connector's config from the
   connector's default template.

2. update

   A category of commands that modifies connector's config. For example, when
   working on Postgres, you may want to introspect the database and import all
   the tables into the connector's config.

   These are just some examples (they are in no way prescriptive of what the
   exact interface should be):

   ```bash
   # 1. Imports the database and updates all the tables that are # not in pg_*
   schemas or information_schema
   # 2. Will also preserve the user written native queries
   <pg-plugin> update-config --replace

   # Same as above but is restricting the introspection to 'public' schema
   <pg-plugin> update-config --include-schema public
   ```

   > [!NOTE]
   > In the case of Postgres connector, the whole `configureOptions` section
   > goes away from the connector's config because that doesn't affect the
   > NDC interface's behaviour and ideally part of update-config's arguments

3. validate

   This should help validate the config for the connector. Note that the
   connector's deployment fails if the config is invalid. However for a user it
   may be easier to have a command that validates the config before it is
   deployed. It isn't strictly necessary for beta but a nice to have.

   In future, you could also have specific commands to validate parts of the
   config such as native queries.

4. watch

   This is basically (for the most common use case) the update command run on loop. 
   This could also be a totally different implementation. This is a convenience 
   command for the CLI to use during `h3 dev` so that the CLI can invoke this once, 
   and terminate it once the `dev` command has ended.  

#### Inputs to the Plugin Invocation

   The main CLI invokes the above commands as sub-processes and passes all the Environment 
   variables specified in the `ConnectorManifest` (for eg. in the case of postgres, it will 
   pass `PG_URL` , etc) to the Plugin. In addition to these Env vars, the main CLI passes
   the following ENV vars:
   - `HASURA_PLUGIN_DDN_PAT` (string)- the PAT token which can be used to make authenticated 
   calls to Hasura Cloud.
   - `HASURA_PLUGIN_DISABLE_TELEMETRY` (boolean string, `true` or `false`) - If the plugins 
   are sending any sort of telemetry back to Hasura, it should be disabled if this is `true`.
   - `HASURA_PLUGIN_INSTANCE_ID` (string) - A UUID for every unique user. Can be used in 
   telemetry.
   - `HASURA_PLUGIN_EXECUTION_ID` (string) - A UUID unique to every invocation of Hasura CLI.
   - `HASURA_PLUGIN_LOG_LEVEL` (string) - Can be one of [these](https://github.com/rs/zerolog?tab=readme-ov-file#leveled-logging) 
   log levels.
   - `HASURA_PLUGIN_CONNECTOR_CONTEXT_PATH` (string) - Fully qualified path to the context
   directory of the connector. 

#### Publishing the Plugin
   The plugin details (name, version, download location, etc., what we call the plugin manifest) 
   should be published to the [cli-plugin-index](https://github.com/hasura/cli-plugins-index) repo. 
   The CLI will track the contents of the `master` branch of this repo and will be able to install
   any version of the published plugin. (Version should be a CalVer of the format `YYYYMMDD` or a SemVer).
   Also, we need plugin binaries for `linux/amd64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64` platforms.
   The name of the plugin must be `hasura-<your-plugin-name>` for it to be recognized in by the CLI. 
   The plugin binary (preferably statically linked) should be pushed as a gzipped tarball and its URI 
   mentioned in the manifest for each platform. The sha256 value in the [manifest](https://github.com/hasura/cli-plugins-index/blob/master/plugins/connector/20240125/manifest.yaml) is the checksum of the gzipped tarball.