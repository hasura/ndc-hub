# Connector metadata types

This project contains the schema of the connector-metadata.json file that is used in the
package.tgz file of a connector, whenever a new version of a connector is released.

It is expected that whenever authors propose a new RFC that changes the schema
of the connector-metadata.json file, they will update this schema file accordingly in the `connectorTypes.ts` file.

The `schema.json` file is the JSON schema of the current `connector-metadata.json` structure.


## How to update the schema

1. Update the `connectorTypes.ts` file with the new schema.
2. Run `npm run generate-schema schema.json` to update the `schema.json` file.

## Github Actions workflow


The repo includes a GitHub Actions workflow that generates the schema and
compares it with the existing schema file.

This ensures that any changes to the schema generation logic are reflected
in the committed schema file.
