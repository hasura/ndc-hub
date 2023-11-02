use schemars::{schema::Schema, JsonSchema};
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug, Clone, JsonSchema)]
#[serde(rename_all = "camelCase")]
#[schemars(title = "SecretValue")]
/// Either a literal string or a reference to a Hasura secret
pub enum SecretValueImpl {
    Value(String),
    StringValueFromSecret(String),
}

/// Use this type to refer to secret strings within a
/// connector's configuration types. For example, a connection
/// string which might contain a password should be configured
/// using this type.
///
/// This marker type indicates that a value should be configured
/// from secrets drawn from a secret store.
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct SecretValue(pub SecretValueImpl);

impl JsonSchema for SecretValue {
    fn schema_name() -> String {
        SecretValueImpl::schema_name()
    }

    fn json_schema(gen: &mut schemars::gen::SchemaGenerator) -> schemars::schema::Schema {
        let mut s = SecretValueImpl::json_schema(gen);
        if let Schema::Object(o) = &mut s {
            if let Some(m) = &mut o.metadata {
                m.id = Some(Self::schema_id().into());
            }
        }
        s
    }

    fn schema_id() -> std::borrow::Cow<'static, str> {
        "https://hasura.io/jsonschemas/SecretValue".into()
    }
}

#[cfg(test)]
mod tests {
    use goldenfile::Mint;
    use schemars::schema_for;
    use std::io::Write;
    use std::path::PathBuf;

    use super::SecretValue;

    #[test]
    pub fn test_json_schema() {
        let test_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("tests");

        let mut mint = Mint::new(test_dir);

        let expected_path = PathBuf::from_iter(["json_schema", "secret_value.jsonschema"]);

        let mut expected = mint.new_goldenfile(expected_path).unwrap();

        let schema = schema_for!(SecretValue);

        write!(
            expected,
            "{}",
            serde_json::to_string_pretty(&schema).unwrap()
        )
        .unwrap();
    }
}
