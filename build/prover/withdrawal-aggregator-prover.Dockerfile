FROM rust:alpine3.20 AS intmax2-rust-build-tools
RUN apk add --no-cache musl-dev
RUN rustup override set nightly

FROM intmax2-rust-build-tools AS intmax2-rust-build-env
WORKDIR /app
ADD . .
RUN cargo build --release --features dummy_proof --bin withdrawal-aggregator-prover

FROM alpine:3.20 AS intmax2-rust-run-env
COPY --from=intmax2-rust-build-env /app/target/release/withdrawal-aggregator-prover /app/withdrawal-aggregator-prover
COPY --from=intmax2-rust-build-env /app/config.toml /app/config.toml
WORKDIR /app

