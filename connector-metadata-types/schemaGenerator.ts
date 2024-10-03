import * as TJS from "typescript-json-schema";
import * as path from "path";
import fs from "fs";

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

function writeSchemaToFile(filePath: string) {
  const schema = generateConnectorMetadataSchema();
  const jsonSchema = JSON.stringify(schema, null, 2);
  fs.writeFileSync(filePath, jsonSchema);
}

if (require.main === module) {
  const outputFile = process.argv[2];
  if (!outputFile) {
    console.error("Please provide an output file name");
    process.exit(1);
  }
  writeSchemaToFile(path.resolve(outputFile));
}
