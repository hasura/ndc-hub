check: format-check build lint test

build:
  cargo build --all-targets --all-features

# re-build on code changes, and run the reference agent each time a build is
# successful
dev:
  mkdir -p ./tmp/empty
  cargo watch \
    -x test \
    -x 'run --bin ndc_hub_example \
    -- serve --configuration ./tmp/empty \
    --otlp-endpoint http://localhost:4317'

format:
  cargo fmt --all
  ! command -v nix > /dev/null || nix fmt

format-check:
  cargo fmt --all --check
  ! command -v nix > /dev/null || nix fmt -- --check .

lint:
  cargo clippy --all-targets --all-features
  cargo machete --with-metadata

lint-apply:
  cargo clippy --fix --all-targets --all-features

test:
  cargo test --all-targets --all-features
