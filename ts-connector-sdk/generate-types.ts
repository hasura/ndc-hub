import { writeFileSync } from "fs";
import { compileFromFile } from "json-schema-to-typescript";

const schemas = [
  "CapabilitiesResponse",
  "SchemaResponse",
  "QueryRequest",
  "QueryResponse",
  "ErrorResponse",
  "ExplainResponse",
  "MutationRequest",
  "MutationResponse",
];

async function generate() {
  console.log("Generating types...");
  for (const schema of schemas) {
    writeFileSync(
      `./schemas/${schema}.d.ts`,
      await compileFromFile(`../api_schemas/generated/${schema}.schema.json`)
    );
  }
  console.log("done!");
}

generate();

// notes to self: dynamic type generation seems to not be working corectly due to generated  type size
// generating the types as a build step seems to work, except for query response. This could be an issue with the query response schema
