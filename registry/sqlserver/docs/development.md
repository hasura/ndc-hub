# SQL Server Connector Development

Instructions for developers who wish to contribute or build upon the connector:

## Prerequistes

1. Install [rustup](https://www.rust-lang.org/tools/install)
2. Install additional tools:
   - `cargo install cargo-watch cargo-insta`
3. Install [just](https://github.com/casey/just)
4. Install [Prettier](https://prettier.io/)
5. Install [Docker](https://www.docker.com/)
6. Install protoc. Here are a few options:
   - `brew install protobuf`
   - `apt-get install protobuf-compiler`
   - `dnf install protobuf-compiler`

## Compile

```sh
cargo build
```

## Create a configuration

Create a configuration in a new directory using the following commands:

1. Initialize a configuration:

   ```sh
   CONNECTION_URI='<sqlserver-connection-string>' cargo run --bin ndc-sqlserver-cli -- --context='<directory>'  initialize
   ```

2. Update the configuration by introspecting the database:

   ```sh
   CONNECTION_URI='<sqlserver-connection-string>' cargo run --bin ndc-sqlserver-cli -- --context='<directory>'  update
   ```

## Run

Run the SQL Server connector with:

```sh
cargo run serve --configuration '<directory>'
```

## Test

To test SQL Server, run:

```sh
cargo test
```

### Write a database execution test

1. Create a new file under `crates/ndc-sqlserver/tests/goldenfiles/<your-test-name>.json`
2. Create a new test in `crates/ndc-sqlserver/tests/query_tests.rs` that looks like this:
    ```rs
    #[tokio::test]
    async fn select_5() {
            let result = run_query("select_5").await;
            insta::assert_json_snapshot!(result);
        }
    ```
3. Run the tests using `just test` or `cargo test`

### Write a SQL translation snapshot test

1. Create a new folder under `crates/query-engine/translation/tests/goldenfiles/<your-test-name>/`
2. Create `request.json` and `tables.json` files in that folder to specify your request
3. Create a new test in `crates/query-engine/translation/tests/tests.rs` that looks like this:
   ```rs
   #[tokio::test]
   async fn select_5() {
       let result = common::test_translation("select_5").await;
       insta::assert_snapshot!(result);
   }
   ```
4. Run the tests using `just test` or `cargo test`

## Linting

Run `just lint` to run clippy linter

run `just lint-apply` to attempt to autofix all linter suggestions

## Formatting

Check your formatting is great with `just format-check`.

Format all Rust code with `just format`.
