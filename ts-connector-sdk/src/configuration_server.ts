import Fastify, { FastifyRequest } from "fastify";
import JSONSchema, { JSONSchemaObject } from "@json-schema-tools/meta-schema";

import { Connector } from "./connector";
import { ConnectorError } from "./error";

import ErrorResponseSchema from "../schemas/ErrorResponse.schema.json";

interface ConfigurationServerOptions {
  port: number;
}

const errorResponses = {
  400: ErrorResponseSchema,
  403: ErrorResponseSchema,
  409: ErrorResponseSchema,
  500: ErrorResponseSchema,
  501: ErrorResponseSchema,
};

export async function start_configuration_server<Configuration, State>(
  connector: Connector<Configuration, State>,
  options: ConfigurationServerOptions
) {
  const server = Fastify({
    logger: true,
  });

  server.get(
    "/",
    {
      schema: {
        response: {
          200: connector.get_configuration_schema(),
          ...errorResponses,
        },
      },
    },
    async function get_schema(request: FastifyRequest): Promise<Configuration> {
      return connector.make_empty_configuration();
    }
  );

  server.post(
    "/",
    {
      schema: {
        body: connector.get_configuration_schema(),
        response: {
          200: connector.get_configuration_schema(),
          ...errorResponses,
        },
      },
    },
    async (
      request: FastifyRequest<{
        Body: Configuration;
      }>
    ): Promise<Configuration> => {
      return connector.update_configuration(
        // type assetion required because Configuration is a generic parameter
        request.body as Configuration
      );
    }
  );

  server.get(
    "/schema",
    {
      schema: {
        response: {
          200: JSONSchema,
          ...errorResponses,
        },
      },
    },
    async (): Promise<JSONSchemaObject> => connector.get_configuration_schema()
  );

  server.post(
    "/validate",
    {
      schema: {
        body: connector.get_configuration_schema(),
        response: {
          200: connector.get_configuration_schema(),
          ...errorResponses,
        },
      },
    },
    async (
      request: FastifyRequest<{ Body: Configuration }>
    ): Promise<Configuration> => {
      return connector.validate_raw_configuration(
        // type assetion required because Configuration is a generic parameter
        request.body as Configuration
      );
    }
  );

  server.setErrorHandler(function (error, request, reply) {
    if (error instanceof ConnectorError) {
      // Log error
      this.log.error(error);
      // Send error response
      reply.status(error.statusCode).send({
        message: error.message,
        details: error.details,
      });
    } else {
      reply.status(500).send({
        message: error.message,
      });
    }
  });

  try {
    await server.listen({ port: options.port });
  } catch (error) {
    server.log.error(error);
    process.exit(1);
  }
}
