import fs from "fs";
import Fastify, { FastifyRequest } from "fastify";

import { Connector } from "./connector";
import { ConnectorError } from "./error";

import CapabilitiesResponseSchema from "../schemas/CapabilitiesResponse.schema.json";
import SchemaResponseSchema from "../schemas/SchemaResponse.schema.json";
import QueryRequestSchema from "../schemas/QueryRequest.schema.json";
import QueryResponseSchema from "../schemas/QueryResponse.schema.json";
import ExplainResponseSchema from "../schemas/ExplainResponse.schema.json";
import MutationRequestSchema from "../schemas/MutationRequest.schema.json";
import MutationResponseSchema from "../schemas/MutationResponse.schema.json";
import ErrorResponseSchema from "../schemas/ErrorResponse.schema.json";
import { CapabilitiesResponse } from "../schemas/CapabilitiesResponse";
import { SchemaResponse } from "../schemas/SchemaResponse";
import { MutationResponse } from "../schemas/MutationResponse";
import { MutationRequest } from "../schemas/MutationRequest";
import { QueryRequest } from "../schemas/QueryRequest";
import { ErrorResponse } from "../schemas/ErrorResponse";

const errorResponses = {
  400: ErrorResponseSchema,
  403: ErrorResponseSchema,
  409: ErrorResponseSchema,
  500: ErrorResponseSchema,
  501: ErrorResponseSchema,
};

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

  server.get(
    "/capabilities",
    {
      schema: {
        response: {
          200: CapabilitiesResponseSchema,
          ...errorResponses,
        },
      },
    },
    (request: FastifyRequest): CapabilitiesResponse => {
      return connector.get_capabilities(configuration);
    }
  );

  server.get("/health", (request): Promise<undefined> => {
    return connector.health_check(configuration, state);
  });

  server.get("/metrics", (request) => {
    return connector.fetch_metrics(configuration, state);
  });

  server.get(
    "/schema",
    {
      schema: {
        response: {
          200: SchemaResponseSchema,
          ...errorResponses,
        },
      },
    },
    (request): SchemaResponse => {
      return connector.get_schema(configuration);
    }
  );

  server.post(
    "/query",
    {
      schema: {
        body: QueryRequestSchema,
        response: {
          200: QueryResponseSchema,
          ...errorResponses,
        },
      },
    },
    async (
      request: FastifyRequest<{
        Body: QueryRequest;
      }>
    ) => {
      return connector.query(configuration, state, request.body);
    }
  );

  server.post(
    "/explain",
    {
      schema: {
        body: QueryRequestSchema,
        response: {
          200: ExplainResponseSchema,
          ...errorResponses,
        },
      },
    },
    (request) => {
      return connector.explain(
        configuration,
        state,
        request.body as QueryRequest
      );
    }
  );

  server.post(
    "/mutation",
    {
      schema: {
        body: MutationRequestSchema,
        response: {
          200: MutationResponseSchema,
          ...errorResponses,
        },
      },
    },
    (
      request: FastifyRequest<{
        Body: MutationRequest;
      }>
    ): Promise<MutationResponse> => {
      return connector.mutation(
        configuration,
        state,
        request.body as MutationRequest
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

function get_configuration<Configuration>(path: string): Configuration {
  const data = fs.readFileSync(path);
  const configuration = JSON.parse(data.toString());
  return configuration as Configuration;
}
