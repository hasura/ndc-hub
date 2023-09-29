dev:
  cargo watch \
    -x 'run --bin ndc_hub_example \
    -- serve --configuration <(echo 'null') \
    --otlp-endpoint http://localhost:4317'
