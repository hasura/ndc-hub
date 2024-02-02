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

   The idea is that the user would specify a specific incantation of the update
   command that cli would then call at a fixed interval when the user invokes
   `h3 dev`.

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

