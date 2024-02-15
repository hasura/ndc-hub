FROM rust:1.70.0-slim-buster AS build

WORKDIR app

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive \
    apt-get install --no-install-recommends --assume-yes \
      lld protobuf-compiler libssl-dev ssh git pkg-config

ENV CARGO_HOME=/app/.cargo
ENV RUSTFLAGS="-C link-arg=-fuse-ld=lld"

COPY Cargo.lock .
COPY ./rust-connector-sdk .

RUN cargo build --release

FROM debian:buster-slim as connector
RUN set -ex; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive \
      apt-get install --no-install-recommends --assume-yes \
      libssl-dev
COPY --from=build /app/target/release/ndc_hub_example ./ndc_hub_example
ENTRYPOINT [ "/ndc_hub_example" ]
CMD [ "serve", "--configuration", "/etc/connector" ]
