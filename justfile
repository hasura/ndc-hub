# re-build on code changes, and run the reference agent each time a build is
# successful
dev:
  cargo watch \
    -x test \
    -x 'run --bin ndc_hub_example \
    -- serve --configuration <(echo 'null') \
    --otlp-endpoint http://localhost:4317'

# reformat everything
format:
  cargo fmt --all

# is everything formatted?
format-check:
  cargo fmt --all --check

# run `clippy` linter
lint *FLAGS:
  cargo clippy {{FLAGS}}

lint-apply *FLAGS:
  cargo clippy --fix {{FLAGS}}



