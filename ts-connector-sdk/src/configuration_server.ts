import { Connector } from "./connector";
import Fastify from "fastify";

interface ConfigurationServerOptions {
  port: number;
}

export async function start_configuration_server<Configuration, State>(
  connector: Connector<Configuration, State>,
  options: ConfigurationServerOptions
) {
  const server = Fastify({
    logger: true,
  });

  server.get("/", async () => connector.make_empty_configuration());

  server.post("/", async (request) => {
    return connector.update_configuration(request.body as Configuration);
  });

  server.get("/schema", async () => connector.get_configuration_schema());

  server.post("/validate", async (request) => {
    return connector.validate_raw_configuration(request.body as Configuration);
  });

  try {
    await server.listen({ port: options.port });
  } catch (error) {
    server.log.error(error);
    process.exit(1);
  }
}
