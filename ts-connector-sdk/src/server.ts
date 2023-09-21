import { MutationRequest } from "../schemas/MutationRequest";
import { QueryRequest } from "../schemas/QueryRequest";
import { Connector } from "./connector";
import Fastify from "fastify";
import fs from "fs";

interface ServerOptions {
  configuration: string;
  port: number;
  serviceTokenSecret: string | null;
  otlpEndpoint: string | null;
  serviceName: string | null;
}

export async function start_server<Configuration, State>(
  connector: Connector<Configuration, State>,
  options: ServerOptions
) {
  const configuration = await get_configuration<Configuration>(
    options.configuration
  );

  const metrics = {}; // todo

  const state = await connector.try_init_state(configuration, metrics);

  const server = Fastify({
    logger: true,
  });

  server.get("/capabilities", (request) => {
    return connector.get_capabilities(configuration);
  });

  server.get("/health", (request) => {
    return connector.health_check(configuration, state);
  });

  server.get("/metrics", (request) => {
    return connector.fetch_metrics(configuration, state);
  });

  server.get("/schema", (request) => {
    return connector.get_schema(configuration);
  });

  server.post("/query", (request) => {
    return connector.query(configuration, state, request.body as QueryRequest);
  });

  server.post("/explain", (request) => {
    return connector.explain(
      configuration,
      state,
      request.body as QueryRequest
    );
  });

  server.post("/mutation", (request) => {
    return connector.mutation(
      configuration,
      state,
      request.body as MutationRequest
    );
  });

  try {
    await server.listen({ port: options.port });
  } catch (error) {
    server.log.error(error);
    process.exit(1);
  }
}

function get_configuration<Configuration>(path: string): Configuration {
  const data = fs.readFileSync(path);
  const configuration = JSON.parse(data.toString());
  return configuration as Configuration;
}
