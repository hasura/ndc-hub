import fs from "fs";
import path from "path";
import { generateConnectorMetadataSchema } from "./schemaGenerator";

function compareSchema() {
  const newSchema = generateConnectorMetadataSchema();
  const existingSchemaPath = path.join(__dirname, "schema.json");

  if (!fs.existsSync(existingSchemaPath)) {
    console.error("Existing schema.json not found. Generating new schema.");
    fs.writeFileSync(existingSchemaPath, JSON.stringify(newSchema, null, 2));
    console.log("New schema has been generated and saved to schema.json");
    return;
  }

  const existingSchema = JSON.parse(
    fs.readFileSync(existingSchemaPath, "utf8"),
  );

  if (JSON.stringify(existingSchema) !== JSON.stringify(newSchema)) {
    console.error(
      "Error: Generated schema does not match the existing schema.",
    );
    process.exit(1);
  } else {
    console.log("Schema validation successful. No changes detected.");
  }
}

compareSchema();
