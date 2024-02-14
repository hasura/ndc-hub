use std::collections::BTreeMap;

use axum::{extract::State, http::StatusCode, response::IntoResponse, Json};
use gdc_rust_types::{
    Aggregate, BinaryArrayComparisonOperator, BinaryComparisonOperator, Capabilities,
    CapabilitiesResponse, ColumnInfo, ColumnSelector, ColumnType, ComparisonCapabilities,
    ComparisonColumn, ComparisonValue, ConfigSchemaResponse, DetailLevel, ErrorResponse,
    ErrorResponseType, ExistsInTable, ExplainResponse, Expression, Field, ForEachRow, FunctionInfo,
    ObjectTypeDefinition, OrderBy, OrderByElement, OrderByRelation, OrderByTarget, OrderDirection,
    Query, QueryRequest, QueryResponse, Relationship, RelationshipType, ResponseFieldValue,
    ResponseRow, ScalarTypeCapabilities, SchemaRequest, SchemaResponse,
    SubqueryComparisonCapabilities, TableInfo, TableRelationships, Target, UnaryComparisonOperator,
};
use indexmap::IndexMap;
use ndc_client::models;
use serde::{Deserialize, Serialize};
use serde_json::json;

use crate::connector::{Connector, ExplainError, QueryError};
use crate::default_main::ServerState;
use crate::json_response::JsonResponse;

pub async fn get_health() -> impl IntoResponse {
    // todo: if source_name and config provided, check if that specific source is healthy
    StatusCode::NO_CONTENT
}

pub async fn get_capabilities<C: Connector>(
    State(state): State<ServerState<C>>,
) -> Result<Json<CapabilitiesResponse>, (StatusCode, Json<ErrorResponse>)> {
    let v3_capabilities = C::get_capabilities().await.into_value().map_err(
        |err: Box<dyn std::error::Error + Send + Sync>| {
            (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse {
                    details: None,
                    message: err.to_string(),
                    r#type: None,
                }),
            )
        },
    )?;
    let v3_schema = C::get_schema(&state.configuration)
        .await
        .and_then(JsonResponse::into_value)
        .map_err(|err| {
            (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse {
                    details: None,
                    message: err.to_string(),
                    r#type: None,
                }),
            )
        })?;

    let scalar_types = IndexMap::from_iter(v3_schema.scalar_types.into_iter().map(
        |(name, scalar_type)| {
            (
                name,
                ScalarTypeCapabilities {
                    aggregate_functions: Some(IndexMap::from_iter(
                        scalar_type.aggregate_functions.into_iter().filter_map(
                            |(function_name, aggregate_function)| match aggregate_function
                                .result_type
                            {
                                models::Type::Named { name } => Some((function_name, name)),
                                models::Type::Nullable { .. } => None,
                                models::Type::Array { .. } => None,
                                models::Type::Predicate { .. } => None,
                            },
                        ),
                    )),
                    comparison_operators: Some(IndexMap::from_iter(
                        scalar_type.comparison_operators.into_iter().filter_map(
                            |(operator_name, comparison_operator)| match comparison_operator {
                                models::ComparisonOperatorDefinition::Equal => {
                                    Some(("equal".to_string(), "equal".to_string()))
                                }
                                models::ComparisonOperatorDefinition::In => {
                                    Some(("in".to_string(), "in".to_string()))
                                }
                                models::ComparisonOperatorDefinition::Custom {
                                    argument_type: models::Type::Named { name },
                                } => Some((operator_name, name)),
                                models::ComparisonOperatorDefinition::Custom {
                                    argument_type: models::Type::Nullable { .. },
                                } => None,
                                models::ComparisonOperatorDefinition::Custom {
                                    argument_type: models::Type::Array { .. },
                                } => None,
                                models::ComparisonOperatorDefinition::Custom {
                                    argument_type: models::Type::Predicate { .. },
                                } => None,
                            },
                        ),
                    )),
                    update_column_operators: None,
                    graphql_type: None,
                },
            )
        },
    ));

    let response = CapabilitiesResponse {
        capabilities: Capabilities {
            comparisons: Some(ComparisonCapabilities {
                subquery: Some(SubqueryComparisonCapabilities {
                    supports_relations: v3_capabilities
                        .capabilities
                        .relationships
                        .as_ref()
                        .map(|capabilities| capabilities.relation_comparisons.is_some()),
                }),
            }),
            data_schema: None,
            datasets: None,
            explain: None,
            interpolated_queries: None,
            licensing: None,
            metrics: None,
            mutations: None,
            queries: None,
            raw: None,
            relationships: None,
            scalar_types: Some(scalar_types),
            subscriptions: None,
            user_defined_functions: None,
            post_schema: Some(json!({})),
        },
        config_schemas: get_openapi_config_schema_response(),
        display_name: None,
        release_name: Some(v3_capabilities.version.to_owned()),
    };

    Ok(Json(response))
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SourceConfig {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub service_token_secret: Option<String>,
}

fn get_openapi_config_schema_response() -> ConfigSchemaResponse {
    // note: we should probably have some config for auth, will do later
    let config_schema_json = json!({
        "type": "object",
        "nullable": false,
        "properties": {
            "service_token_secret": {
                "title": "Service Token Secret",
                "description": "Service Token Secret, required if your connector is configured with a secret.",
                "nullable": true,
                "type": "string"
            }
        },
        "required": []
    });

    ConfigSchemaResponse {
        config_schema: serde_json::from_value(config_schema_json)
            .expect("json value should be valid OpenAPI schema"),
        other_schemas: serde_json::from_str("{}").expect("static string should be valid json"),
    }
}

pub async fn post_schema<C: Connector>(
    State(state): State<ServerState<C>>,
    request: Option<Json<SchemaRequest>>,
) -> Result<Json<SchemaResponse>, (StatusCode, Json<ErrorResponse>)> {
    let v3_schema = C::get_schema(&state.configuration)
        .await
        .and_then(JsonResponse::into_value)
        .map_err(|err| {
            (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse {
                    details: None,
                    message: err.to_string(),
                    r#type: None,
                }),
            )
        })?;
    let schema = map_schema(v3_schema).map_err(|err| (StatusCode::BAD_REQUEST, Json(err)))?;

    let schema = if let Some(request) = request {
        let SchemaResponse {
            object_types,
            tables,
            functions,
        } = schema;

        let tables = if let Some(requested_tables) = request
            .filters
            .as_ref()
            .and_then(|filters| filters.only_tables.as_ref())
        {
            tables
                .into_iter()
                .filter(|table| {
                    requested_tables
                        .iter()
                        .any(|requested_table| requested_table == &table.name)
                })
                .collect()
        } else {
            tables
        };

        let tables = match request.detail_level {
            Some(DetailLevel::BasicInfo) => tables
                .into_iter()
                .map(|table| TableInfo {
                    columns: None,
                    deletable: None,
                    description: None,
                    foreign_keys: None,
                    insertable: None,
                    name: table.name,
                    primary_key: None,
                    r#type: table.r#type,
                    updatable: None,
                })
                .collect(),
            _ => tables,
        };

        let functions = if let Some(requested_functions) = request
            .filters
            .as_ref()
            .and_then(|filters| filters.only_functions.as_ref())
        {
            functions.map(|functions| {
                functions
                    .into_iter()
                    .filter(|function| {
                        requested_functions
                            .iter()
                            .any(|requested_function| requested_function == &function.name)
                    })
                    .collect()
            })
        } else {
            functions
        };

        let functions = match request.detail_level {
            Some(DetailLevel::BasicInfo) => functions.map(|functions| {
                functions
                    .into_iter()
                    .map(|function| FunctionInfo {
                        args: None,
                        description: None,
                        name: function.name,
                        response_cardinality: None,
                        returns: None,
                        r#type: function.r#type,
                    })
                    .collect()
            }),
            _ => functions,
        };

        SchemaResponse {
            object_types,
            tables,
            functions,
        }
    } else {
        schema
    };

    Ok(Json(schema))
}

fn map_schema(schema: models::SchemaResponse) -> Result<SchemaResponse, ErrorResponse> {
    let tables = schema
        .collections
        .iter()
        .map(|collection| {
            let table_type = schema
                .object_types
                .get(&collection.collection_type)
                .ok_or_else(|| ErrorResponse {
                    details: None,
                    message: format!(
                        "Could not find type {} for table {}",
                        collection.collection_type, collection.name
                    ),
                    r#type: Some(ErrorResponseType::UncaughtError),
                })?;
            let columns = table_type
                .fields
                .iter()
                .map(|(field_name, field_info)| {
                    Ok(ColumnInfo {
                        name: field_name.to_owned(),
                        r#type: get_field_type(&field_info.r#type, &schema),
                        nullable: matches!(field_info.r#type, models::Type::Nullable { .. }),
                        description: field_info.description.to_owned(),
                        insertable: None,
                        updatable: None,
                        value_generated: None,
                    })
                })
                .collect::<Result<Vec<_>, _>>()?;
            Ok(TableInfo {
                name: vec![collection.name.to_owned()],
                description: collection.description.to_owned(),
                insertable: None,
                updatable: None,
                deletable: None,
                primary_key: None,
                foreign_keys: None,
                r#type: None,
                columns: Some(columns),
            })
        })
        .collect::<Result<Vec<_>, _>>()?;

    let object_types = schema
        .object_types
        .iter()
        .map(|(object_name, object_definition)| {
            Ok(ObjectTypeDefinition {
                name: object_name.to_owned(),
                description: object_definition.description.to_owned(),
                columns: object_definition
                    .fields
                    .iter()
                    .map(|(field_name, field_definition)| ColumnInfo {
                        description: field_definition.description.to_owned(),
                        insertable: None,
                        name: field_name.to_owned(),
                        nullable: matches!(field_definition.r#type, models::Type::Nullable { .. }),
                        r#type: get_field_type(&field_definition.r#type, &schema),
                        updatable: None,
                        value_generated: None,
                    })
                    .collect(),
            })
        })
        .collect::<Result<Vec<_>, _>>()?;

    Ok(SchemaResponse {
        tables,
        object_types: Some(object_types),
        functions: None,
    })
}

fn get_field_type(column_type: &models::Type, schema: &models::SchemaResponse) -> ColumnType {
    match column_type {
        models::Type::Named { name } => {
            if schema.object_types.contains_key(name) {
                ColumnType::ColumnTypeNonScalar(gdc_rust_types::ColumnTypeNonScalar::Object {
                    name: name.to_owned(),
                })
            } else {
                // silently assuming scalar if not object type
                ColumnType::Scalar(name.to_owned())
            }
        }
        models::Type::Nullable { underlying_type } => get_field_type(underlying_type, schema),
        models::Type::Array { element_type } => {
            ColumnType::ColumnTypeNonScalar(gdc_rust_types::ColumnTypeNonScalar::Array {
                element_type: Box::new(get_field_type(element_type, schema)),
                nullable: matches!(**element_type, models::Type::Nullable { .. }),
            })
        }
        models::Type::Predicate { .. } => todo!(),
    }
}

pub async fn post_query<C: Connector>(
    State(state): State<ServerState<C>>,
    Json(request): Json<QueryRequest>,
) -> Result<Json<QueryResponse>, (StatusCode, Json<ErrorResponse>)> {
    let request = map_query_request(request).map_err(|err| (StatusCode::BAD_REQUEST, Json(err)))?;
    let response = C::query(&state.configuration, &state.state, request)
        .await
        .and_then(JsonResponse::into_value)
        .map_err(|err| match err {
            QueryError::InvalidRequest(message)
            | QueryError::UnsupportedOperation(message)
            | QueryError::UnprocessableContent(message) => (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    details: None,
                    message,
                    r#type: None,
                }),
            ),
            QueryError::Other(err) => (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    details: None,
                    message: err.to_string(),
                    r#type: None,
                }),
            ),
        })?;
    Ok(Json(map_query_response(response)))
}

pub async fn post_explain<C: Connector>(
    State(state): State<ServerState<C>>,
    Json(request): Json<QueryRequest>,
) -> Result<Json<ExplainResponse>, (StatusCode, Json<ErrorResponse>)> {
    let v2_ir_json = serde_json::to_string(&request).map_err(|err| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ErrorResponse {
                details: None,
                message: format!("Error serializing v2 IR to JSON: {}", err),
                r#type: None,
            }),
        )
    })?;
    let request = map_query_request(request).map_err(|err| (StatusCode::BAD_REQUEST, Json(err)))?;

    let v3_ir_json = serde_json::to_string(&request).map_err(|err| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ErrorResponse {
                details: None,
                message: format!("Error serializing v3 IR to JSON: {}", err),
                r#type: None,
            }),
        )
    })?;
    let response = C::query_explain(&state.configuration, &state.state, request.clone())
        .await
        .and_then(JsonResponse::into_value)
        .map_err(|err| match err {
            ExplainError::InvalidRequest(message)
            | ExplainError::UnsupportedOperation(message)
            | ExplainError::UnprocessableContent(message) => (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    details: None,
                    message,
                    r#type: None,
                }),
            ),
            ExplainError::Other(err) => (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    details: None,
                    message: err.to_string(),
                    r#type: None,
                }),
            ),
        })?;

    let response = ExplainResponse {
        lines: vec![
            "v2 IR".to_string(),
            v2_ir_json,
            "v3 IR".to_string(),
            v3_ir_json,
        ]
        .into_iter()
        .chain(
            response
                .details
                .into_iter()
                .map(|(key, value)| format!("{key}: {value}")),
        )
        .collect(),
        query: "".to_string(),
    };
    Ok(Json(response))
}

fn map_query_request(request: QueryRequest) -> Result<models::QueryRequest, ErrorResponse> {
    let QueryRequest {
        foreach,
        target,
        relationships,
        query,
        interpolated_queries: _,
    } = request;

    let foreach_expr = foreach
        .as_ref()
        .and_then(|foreach| foreach.first())
        .and_then(|first_row| {
            let mut expressions: Vec<_> = first_row
                .keys()
                .map(|key| models::Expression::BinaryComparisonOperator {
                    column: models::ComparisonTarget::Column {
                        name: key.to_owned(),
                        path: vec![],
                    },
                    operator: "equal".to_string(),
                    value: models::ComparisonValue::Variable {
                        name: key.to_owned(),
                    },
                })
                .collect();

            if expressions.len() > 1 {
                Some(models::Expression::And { expressions })
            } else {
                expressions.pop()
            }
        });

    let variables = foreach.map(|foreach| {
        foreach
            .into_iter()
            .map(|map| BTreeMap::from_iter(map.into_iter().map(|(key, value)| (key, value.value))))
            .collect()
    });

    let (collection, arguments) = get_collection_and_arguments(&target)?;

    let collection_relationships = BTreeMap::from_iter(
        relationships
            .iter()
            .map(|source_table| {
                let collection = get_name(&source_table.source_table)?;
                source_table
                    .relationships
                    .iter()
                    .map(move |(relationship_name, relationship_info)| {
                        let Relationship {
                            column_mapping,
                            relationship_type,
                            target,
                        } = relationship_info;
                        let (target_collection, arguments) =
                            get_collection_and_relationship_arguments(target)?;
                        Ok((
                            format!("{}.{}", collection, relationship_name),
                            models::Relationship {
                                column_mapping: BTreeMap::from_iter(
                                    column_mapping.clone().into_iter(),
                                ),
                                relationship_type: match relationship_type {
                                    RelationshipType::Object => models::RelationshipType::Object,
                                    RelationshipType::Array => models::RelationshipType::Array,
                                },
                                target_collection,
                                arguments,
                            },
                        ))
                    })
                    .collect::<Result<Vec<_>, _>>()
            })
            .collect::<Result<Vec<_>, _>>()?
            .into_iter()
            .flatten(),
    );

    Ok(models::QueryRequest {
        collection: collection.clone(),
        arguments,
        variables,
        query: map_query(query, &collection, &relationships, foreach_expr)?,
        collection_relationships,
    })
}

fn map_query(
    query: Query,
    collection: &String,
    relationships: &Vec<TableRelationships>,
    foreach_expr: Option<models::Expression>,
) -> Result<models::Query, ErrorResponse> {
    let Query {
        aggregates,
        aggregates_limit,
        fields,
        limit,
        offset,
        order_by,
        r#where,
    } = query;

    let order_by = order_by
        .map(|order_by| {
            let OrderBy {
                elements,
                relations,
            } = order_by;
            Ok(models::OrderBy {
                elements: elements
                    .into_iter()
                    .map(|element| {
                        let OrderByElement {
                            order_direction,
                            target,
                            target_path,
                        } = element;

                        let element = models::OrderByElement {
                            order_direction: match order_direction {
                                OrderDirection::Asc => models::OrderDirection::Asc,
                                OrderDirection::Desc => models::OrderDirection::Desc,
                            },
                            target: match target {
                                OrderByTarget::StarCountAggregate {} => {
                                    models::OrderByTarget::StarCountAggregate {
                                        path: map_order_by_path(
                                            target_path,
                                            relations.to_owned(),
                                            collection,
                                            relationships,
                                        )?,
                                    }
                                }
                                OrderByTarget::SingleColumnAggregate {
                                    column,
                                    function,
                                    result_type: _,
                                } => models::OrderByTarget::SingleColumnAggregate {
                                    column,
                                    function,
                                    path: map_order_by_path(
                                        target_path,
                                        relations.to_owned(),
                                        collection,
                                        relationships,
                                    )?,
                                },
                                OrderByTarget::Column { column } => models::OrderByTarget::Column {
                                    name: get_col_name(&column)?,
                                    path: map_order_by_path(
                                        target_path,
                                        relations.to_owned(),
                                        collection,
                                        relationships,
                                    )?,
                                },
                            },
                        };
                        Ok(element)
                    })
                    .collect::<Result<Vec<_>, _>>()?,
            })
        })
        .transpose()?;

    let aggregates = aggregates.map(|aggregates| {
        IndexMap::from_iter(aggregates.into_iter().map(|(key, aggregate)| {
            (
                key,
                match aggregate {
                    Aggregate::ColumnCount { column, distinct } => {
                        models::Aggregate::ColumnCount { column, distinct }
                    }
                    Aggregate::SingleColumn {
                        column,
                        function,
                        result_type: _,
                    } => models::Aggregate::SingleColumn { column, function },
                    Aggregate::StarCount {} => models::Aggregate::StarCount {},
                },
            )
        }))
    });
    let fields = fields
        .map(|fields| {
            let fields = fields
                .into_iter()
                .map(|(key, field)| {
                    Ok((
                        key,
                        match field {
                            Field::Column {
                                column,
                                column_type: _,
                            } => models::Field::Column {
                                column,
                                fields: None,
                            },
                            Field::Relationship {
                                query,
                                relationship,
                            } => {
                                let (target_collection, arguments) =
                                    get_relationship_collection_arguments(
                                        collection,
                                        &relationship,
                                        relationships,
                                    )?;

                                models::Field::Relationship {
                                    query: Box::new(map_query(
                                        query,
                                        &target_collection,
                                        relationships,
                                        None,
                                    )?),
                                    relationship: format!("{}.{}", collection, relationship),
                                    arguments,
                                }
                            }
                            Field::Object { .. } => {
                                return Err(ErrorResponse {
                                    details: None,
                                    message: "Object fields not supported".to_string(),
                                    r#type: None,
                                })
                            }
                            Field::Array { .. } => {
                                return Err(ErrorResponse {
                                    details: None,
                                    message: "Array fields not supported".to_string(),
                                    r#type: None,
                                })
                            }
                        },
                    ))
                })
                .collect::<Result<Vec<(String, models::Field)>, _>>()?
                .into_iter();
            Ok(IndexMap::from_iter(fields))
        })
        .transpose()?;

    let applicable_limit = match (limit, aggregates_limit) {
        (None, None) => None,
        (None, Some(aggregates_limit)) => {
            if fields.is_none() {
                Some(aggregates_limit)
            } else {
                return Err(ErrorResponse {
                    details: None,
                    message:
                        "Setting limit for aggregates when fields also requested is not supported"
                            .to_string(),
                    r#type: None,
                });
            }
        }
        (Some(limit), None) => {
            if aggregates.is_none() {
                Some(limit)
            } else {
                return Err(ErrorResponse {
                    details: None,
                    message:
                        "Setting limit for fields when aggregates also requested is not supported"
                            .to_string(),
                    r#type: None,
                });
            }
        }
        (Some(_), Some(_)) => {
            return Err(ErrorResponse {
                details: None,
                message: "Different limits for aggregates and fields not supported".to_string(),
                r#type: None,
            })
        }
    };

    let limit = applicable_limit
        .map(|limit| {
            limit.try_into().map_err(|_| ErrorResponse {
                details: None,
                message: "Limit must be valid u32".to_string(),
                r#type: None,
            })
        })
        .transpose()?;

    let offset = offset
        .map(|offset| {
            offset.try_into().map_err(|_| ErrorResponse {
                details: None,
                message: "Offset must be valid u32".to_string(),
                r#type: None,
            })
        })
        .transpose()?;

    let predicate = r#where
        .map(|r#where| map_expression(&r#where, collection, relationships))
        .transpose()?;

    let predicate = match (predicate, foreach_expr) {
        (None, None) => None,
        (None, Some(foreach_expr)) => Some(foreach_expr),
        (Some(predicate), None) => Some(predicate),
        (Some(predicate), Some(foreach_expr)) => Some(models::Expression::And {
            expressions: vec![predicate, foreach_expr],
        }),
    };

    Ok(models::Query {
        aggregates,
        fields,
        limit,
        offset,
        order_by,
        predicate,
    })
}

fn map_order_by_path(
    path: Vec<String>,
    relations: IndexMap<String, OrderByRelation>,
    collection: &String,
    relationships: &Vec<TableRelationships>,
) -> Result<Vec<models::PathElement>, ErrorResponse> {
    let mut mapped_path: Vec<models::PathElement> = vec![];

    let mut relations = relations;
    let mut source_table = collection.to_owned();
    for segment in path {
        let relation = relations.get(&segment).ok_or_else(|| ErrorResponse {
            details: None,
            message: format!("could not find order by relationship for path segment {segment}"),
            r#type: None,
        })?;

        let (target_table, arguments) =
            get_relationship_collection_arguments(&source_table, &segment, relationships)?;

        mapped_path.push(models::PathElement {
            relationship: format!("{}.{}", source_table, segment),
            arguments,
            predicate: if let Some(predicate) = &relation.r#where {
                Some(Box::new(map_expression(
                    predicate,
                    &target_table,
                    relationships,
                )?))
            } else {
                None
            },
        });

        source_table = target_table;
        relations = relation.subrelations.to_owned();
    }

    Ok(mapped_path)
}

fn get_relationship_collection_arguments(
    source_table_name: &str,
    relationship: &str,
    table_relationships: &[TableRelationships],
) -> Result<(String, BTreeMap<String, models::RelationshipArgument>), ErrorResponse> {
    let source_table = table_relationships
        .iter()
        .find(
            |table_relationships| matches!(table_relationships.source_table.as_slice(), [name] if source_table_name == name),
        )
        .ok_or_else(|| ErrorResponse {
            details: None,
            message: format!("Could not find table {source_table_name} in relationships"),
            r#type: None,
        })?;

    let relationship = source_table
        .relationships
        .get(relationship)
        .ok_or_else(|| ErrorResponse {
            details: None,
            message: format!(
                "Could not find relationship {relationship} in table {source_table_name}"
            ),
            r#type: None,
        })?;

    get_collection_and_relationship_arguments(&relationship.target)
}

fn map_expression(
    expression: &Expression,
    collection: &str,
    relationships: &Vec<TableRelationships>,
) -> Result<models::Expression, ErrorResponse> {
    Ok(match expression {
        Expression::And { expressions } => models::Expression::And {
            expressions: expressions
                .iter()
                .map(|expression| map_expression(expression, collection, relationships))
                .collect::<Result<Vec<_>, _>>()?,
        },
        Expression::Or { expressions } => models::Expression::Or {
            expressions: expressions
                .iter()
                .map(|expression| map_expression(expression, collection, relationships))
                .collect::<Result<Vec<_>, _>>()?,
        },
        Expression::Not { expression } => models::Expression::Not {
            expression: Box::new(map_expression(expression, collection, relationships)?),
        },
        Expression::ApplyUnaryComparison { column, operator } => {
            models::Expression::UnaryComparisonOperator {
                column: map_comparison_column(column)?,
                operator: match operator {
                    UnaryComparisonOperator::IsNull => models::UnaryComparisonOperator::IsNull,
                    UnaryComparisonOperator::Other(operator) => {
                        return Err(ErrorResponse {
                            details: None,
                            message: format!("Unknown unary comparison operator {operator}"),
                            r#type: None,
                        })
                    }
                },
            }
        }
        Expression::ApplyBinaryComparison {
            column,
            operator,
            value,
        } => models::Expression::BinaryComparisonOperator {
            column: map_comparison_column(column)?,
            operator: match operator {
                BinaryComparisonOperator::LessThan => "less_than".to_string(),
                BinaryComparisonOperator::LessThanOrEqual => "less_than_or_equal".to_string(),
                BinaryComparisonOperator::Equal => "equal".to_string(),
                BinaryComparisonOperator::GreaterThan => "greater_than".to_string(),
                BinaryComparisonOperator::GreaterThanOrEqual => "greater_than_or_equal".to_string(),
                BinaryComparisonOperator::Other(operator) => operator.to_owned(),
            },
            value: match value {
                ComparisonValue::Scalar {
                    value,
                    value_type: _,
                } => models::ComparisonValue::Scalar {
                    value: value.clone(),
                },
                ComparisonValue::Column { column } => models::ComparisonValue::Column {
                    column: map_comparison_column(column)?,
                },
            },
        },
        Expression::ApplyBinaryArrayComparison {
            column,
            operator,
            value_type: _,
            values,
        } => models::Expression::BinaryComparisonOperator {
            column: map_comparison_column(column)?,
            operator: match operator {
                BinaryArrayComparisonOperator::In => "in".to_string(),
                BinaryArrayComparisonOperator::Other(operator) => {
                    return Err(ErrorResponse {
                        details: None,
                        message: format!("Unknown binary array comparison operator {operator}"),
                        r#type: None,
                    })
                }
            },
            value: models::ComparisonValue::Scalar {
                value: serde_json::to_value(values).unwrap(),
            },
        },
        Expression::Exists { in_table, r#where } => match in_table {
            ExistsInTable::Unrelated { table } => models::Expression::Exists {
                in_collection: models::ExistsInCollection::Unrelated {
                    collection: get_name(table)?,
                    arguments: BTreeMap::new(),
                },
                predicate: Some(Box::new(map_expression(
                    r#where,
                    &get_name(table)?,
                    relationships,
                )?)),
            },
            ExistsInTable::Related { relationship } => {
                let (target_table, arguments) =
                    get_relationship_collection_arguments(collection, relationship, relationships)?;

                models::Expression::Exists {
                    in_collection: models::ExistsInCollection::Related {
                        relationship: format!("{}.{}", collection, relationship),
                        arguments,
                    },
                    predicate: Some(Box::new(map_expression(
                        r#where,
                        &target_table,
                        relationships,
                    )?)),
                }
            }
        },
    })
}

fn map_comparison_column(
    column: &ComparisonColumn,
) -> Result<models::ComparisonTarget, ErrorResponse> {
    match &column.path.as_deref() {
        Some([]) | None => Ok(models::ComparisonTarget::Column {
                    name: get_col_name(&column.name)?,
                    path: vec![],
                }),
        Some([p]) if p == "$" => Ok(models::ComparisonTarget::RootCollectionColumn {
                    name: get_col_name(&column.name)?,
                }),
        Some(path) => Err(ErrorResponse {
            details: None,
            message: format!("Valid values for path are empty array, or array with $ reference to closest query target. Got {}", path.join(".")),
            r#type: None,
        }),
    }
}

fn map_query_response(models::QueryResponse(response): models::QueryResponse) -> QueryResponse {
    if response.len() == 1 {
        QueryResponse::Single(get_reponse_row(
            response
                .into_iter()
                .next()
                .expect("we just checked there is exactly least one element"),
        ))
    } else {
        QueryResponse::ForEach {
            rows: response
                .into_iter()
                .map(|row| ForEachRow {
                    query: get_reponse_row(row),
                })
                .collect(),
        }
    }
}

fn get_reponse_row(row: models::RowSet) -> ResponseRow {
    ResponseRow {
        aggregates: row.aggregates,
        rows: row.rows.map(|rows| {
            rows.into_iter()
                .map(|row| {
                    IndexMap::from_iter(row.into_iter().map(
                        |(alias, models::RowFieldValue(value))| {
                            (alias, ResponseFieldValue::Column(value))
                        },
                    ))
                })
                .collect()
        }),
    }
}

fn get_collection_and_arguments(
    target: &Target,
) -> Result<(String, BTreeMap<String, models::Argument>), ErrorResponse> {
    match target {
        Target::Table { name } => Ok((get_name(name)?, BTreeMap::new())),
        Target::Interpolated { .. } => Err(ErrorResponse {
            details: None,
            message: "Interpolated queries not supported".to_string(),
            r#type: None,
        }),
        Target::Function { name, arguments } => Ok((
            get_name(name)?,
            BTreeMap::from_iter(arguments.iter().map(|argument| match argument {
                gdc_rust_types::FunctionRequestArgument::Named { name, value } => (
                    name.to_owned(),
                    models::Argument::Literal {
                        value: match value {
                            gdc_rust_types::ArgumentValue::Scalar {
                                value,
                                value_type: _,
                            } => value.to_owned(),
                        },
                    },
                ),
            })),
        )),
    }
}

fn get_collection_and_relationship_arguments(
    target: &Target,
) -> Result<(String, BTreeMap<String, models::RelationshipArgument>), ErrorResponse> {
    match target {
        Target::Table { name } => Ok((get_name(name)?, BTreeMap::new())),
        Target::Interpolated { .. } => Err(ErrorResponse {
            details: None,
            message: "Interpolated queries not supported".to_string(),
            r#type: None,
        }),
        Target::Function { name, arguments } => Ok((
            get_name(name)?,
            BTreeMap::from_iter(arguments.iter().map(|argument| match argument {
                gdc_rust_types::FunctionRequestArgument::Named { name, value } => (
                    name.to_owned(),
                    models::RelationshipArgument::Literal {
                        value: match value {
                            gdc_rust_types::ArgumentValue::Scalar {
                                value,
                                value_type: _,
                            } => value.to_owned(),
                        },
                    },
                ),
            })),
        )),
    }
}

fn get_name(target: &Vec<String>) -> Result<String, ErrorResponse> {
    match target.as_slice() {
        [name] => Ok(name.to_owned()),
        _ => Err(ErrorResponse {
            details: None,
            message: format!(
                "Expected function name to be array with exacly one string member, got {}",
                target.join(".")
            ),
            r#type: None,
        }),
    }
}

fn get_col_name(column: &ColumnSelector) -> Result<String, ErrorResponse> {
    match column {
        ColumnSelector::Compound(name) => Err(ErrorResponse {
            details: None,
            message: format!(
                "Compound column selectors not supported, got {}",
                name.join(".")
            ),
            r#type: None,
        }),
        ColumnSelector::Name(name) => Ok(name.to_owned()),
    }
}
