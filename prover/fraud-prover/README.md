# Fraud-prover

## Development

```sh
# env
cp .env.example .env

# run app
RUST_LOG=debug cargo run -r --features dummy_proof
```

## APIs

```sh
FRAUD_PROVER_URL=http://localhost:8080

# heath heck
curl $FRAUD_PROVER_URL/health | jq
```

### FRAUD_PROOF

```sh
# generate proof
curl -X POST -d '{ "id": "1", "challenger":"0x9Fa732B331a5455125c57f9db2E54E03c1CbbA5E", "fraudProof":"'$(base64 --input data/fraud_proof_0x436d3f984fe2d267a6cf8bec2cd062473dd56cec08540115c430474c25f9be4e.bin)'" }' -H "Content-Type: application/json" $FRAUD_PROVER_URL/proof/wrapper | jq

# get proof
curl $FRAUD_PROVER_URL/proof/fraud/2 | jq

# get proofs
curl "$FRAUD_PROVER_URL/proofs/wrapper?ids[]=1&ids[]=2" | jq
```

## Docker

```sh
docker run -d \
  --name prover-redis \
  --hostname redis \
  --restart always \
  -p 6379:6379 \
  -v redisdata:/data \
  redis:7.2.3 \
  /bin/sh -c "redis-server --requirepass password"
```
