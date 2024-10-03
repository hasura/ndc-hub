import * as TJS from "typescript-json-schema";
import * as path from "path";

export function generateConnectorMetadataSchema(): TJS.Definition {
  // Specify the path to the file containing ConnectorMetadataDefinition
  const filePath = path.join(__dirname, "connectorTypes.ts");

  // Compile the TypeScript program
  const program = TJS.getProgramFromFiles([filePath], {
    strictNullChecks: true,
  });

  // Generate the schema
  const schema = TJS.generateSchema(program, "ConnectorMetadataDefinition", {
    required: true,
    ref: false,
  });

  if (!schema) {
    throw new Error(
      "Failed to generate schema for ConnectorMetadataDefinition",
    );
  }

  return schema;
}

// If you want to run this file directly and output the schema
if (require.main === module) {
  const schema = generateConnectorMetadataSchema();
  console.log(JSON.stringify(schema, null, 2));
}
