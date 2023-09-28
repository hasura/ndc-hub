import { CapabilitiesResponse } from "../schemas/CapabilitiesResponse";
import { QueryRequest } from "../schemas/QueryRequest";
import { QueryResponse } from "../schemas/QueryResponse";
import { SchemaResponse } from "../schemas/SchemaResponse";
import { ExplainResponse } from "../schemas/ExplainResponse";
import { MutationRequest } from "../schemas/MutationRequest";
import { MutationResponse } from "../schemas/MutationResponse";
import { JSONSchemaObject } from "@json-schema-tools/meta-schema";

export interface Connector<Configuration, State> {
  /**
   * Return jsonschema for the configuration for this connector
   */
  get_configuration_schema(): JSONSchemaObject;
  /**
   * Return any read regions defined in the connector's configuration
   * @param configuration
   */
  get_read_regions(configuration: Configuration): string[];
  /**
   * Return any write regions defined in the connector's configuration
   * @param configuration
   */
  get_write_regions(configuration: Configuration): string[];

  make_empty_configuration(): Configuration;
  update_configuration(configuration: Configuration): Promise<Configuration>;
  /**
   * Validate the raw configuration provided by the user,
   * returning a configuration error or a validated [`Connector::Configuration`].
   * @param configuration
   */
  validate_raw_configuration(
    configuration: Configuration
  ): Promise<Configuration>;

  /**
   * Initialize the connector's in-memory state.
   *
   * For example, any connection pools, prepared queries,
   * or other managed resources would be allocated here.
   *
   * In addition, this function should register any
   * connector-specific metrics with the metrics registry.
   * @param configuration
   * @param metrics
   */
  try_init_state(
    configuration: Configuration,
    metrics: unknown
  ): Promise<State>;
  /**
   *
   * Update any metrics from the state
   *
   * Note: some metrics can be updated directly, and do not
   * need to be updated here. This function can be useful to
   * query metrics which cannot be updated directly, e.g.
   * the number of idle connections in a connection pool
   * can be polled but not updated directly.
   * @param configuration
   * @param state
   */
  fetch_metrics(configuration: Configuration, state: State): Promise<undefined>;
  /**
   * Check the health of the connector.
   *
   * For example, this function should check that the connector
   * is able to reach its data source over the network.
   *
   * Should throw if the check fails, else resolve
   * @param configuration
   * @param state
   */
  health_check(configuration: Configuration, state: State): Promise<undefined>;

  /**
   * Get the connector's capabilities.
   *
   * This function implements the [capabilities endpoint](https://hasura.github.io/ndc-spec/specification/capabilities.html)
   * from the NDC specification.
   *
   * This function should be syncronous
   * @param configuration
   */
  get_capabilities(configuration: Configuration): CapabilitiesResponse;

  /**
   * Get the connector's schema.
   *
   * This function implements the [schema endpoint](https://hasura.github.io/ndc-spec/specification/schema/index.html)
   * from the NDC specification.
   *
   * This function should resolve from the configuration alone, and is thus syncronous
   * @param configuration
   */
  get_schema(configuration: Configuration): SchemaResponse;

  /**
   * Explain a query by creating an execution plan
   *
   * This function implements the [explain endpoint](https://hasura.github.io/ndc-spec/specification/explain.html)
   * from the NDC specification.
   * @param configuration
   * @param state
   * @param request
   */
  explain(
    configuration: Configuration,
    state: State,
    request: QueryRequest
  ): Promise<ExplainResponse>;

  /**
   * Execute a mutation
   *
   * This function implements the [mutation endpoint](https://hasura.github.io/ndc-spec/specification/mutations/index.html)
   * from the NDC specification.
   * @param configuration
   * @param state
   * @param request
   */
  mutation(
    configuration: Configuration,
    state: State,
    request: MutationRequest
  ): Promise<MutationResponse>;

  /**
   * Execute a query
   *
   * This function implements the [query endpoint](https://hasura.github.io/ndc-spec/specification/queries/index.html)
   * from the NDC specification.
   * @param configuration
   * @param state
   * @param request
   */
  query(
    configuration: Configuration,
    state: State,
    request: QueryRequest
  ): Promise<QueryResponse>;
}
