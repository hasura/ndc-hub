check: format-check build lint test

build:
  cargo build --all-targets --all-features

# re-build on code changes, and run the reference agent each time a build is
# successful
dev:
  cargo watch \
    -x test \
    -x 'run --bin ndc_hub_example \
    -- serve --configuration <(echo 'null') \
    --otlp-endpoint http://localhost:4317'

format:
  cargo fmt --all

format-check:
  cargo fmt --all --check

lint:
  cargo clippy --all-targets --all-features

lint-apply:
  cargo clippy --fix --all-targets --all-features

test:
  cargo test --all-targets --all-features
