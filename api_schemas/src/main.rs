use ndc_client::models::{
    CapabilitiesResponse, ErrorResponse, ExplainResponse, MutationRequest, MutationResponse,
    QueryRequest, QueryResponse, SchemaResponse,
};
use schemars::{schema_for, JsonSchema};
use std::{error::Error, fs};

fn main() -> Result<(), Box<dyn Error>> {
    print!("Generating Schemas...");

    generate_schema::<CapabilitiesResponse>("CapabilitiesResponse")?;
    generate_schema::<SchemaResponse>("SchemaResponse")?;
    generate_schema::<QueryRequest>("QueryRequest")?;
    generate_schema::<QueryResponse>("QueryResponse")?;
    generate_schema::<MutationRequest>("MutationRequest")?;
    generate_schema::<MutationResponse>("MutationResponse")?;
    generate_schema::<ExplainResponse>("ExplainResponse")?;
    generate_schema::<ErrorResponse>("ErrorResponse")?;

    println!("done!");

    Ok(())
}

fn generate_schema<T: JsonSchema>(name: &str) -> Result<(), Box<dyn Error>> {
    fs::write(
        format!("./api_schemas/generated/{name}.schema.json"),
        serde_json::to_string_pretty(&schema_for!(T))?,
    )?;
    Ok(())
}
