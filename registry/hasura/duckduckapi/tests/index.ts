// From original index.ts
import { start } from "@hasura/ndc-duckduckapi";
import { makeConnector, duckduckapi } from "@hasura/ndc-duckduckapi";
import * as path from "path";

const connectorConfig: duckduckapi = {
  dbSchema: `
    CREATE TABLE IF NOT EXISTS messages (
      id INTEGER PRIMARY KEY,
      text TEXT NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
  `,
  functionsFilePath: path.resolve(__dirname, "./functions.ts"),
};

(async () => {
  const connector = await makeConnector(connectorConfig);
  start(connector);
})();
