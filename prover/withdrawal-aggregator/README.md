# withdraw-aggregator-prover

## Development

```sh
# env
cp .env.example .env

# run app
RUST_LOG=debug cargo run -r --features dummy_proof
```

## APIs

```sh
WITHDRAWAL_PROVER_URL=http://localhost:8080

# heath heck
curl $WITHDRAWAL_PROVER_URL/health | jq
```

### Withdrawal

```sh
# generate proof
curl -X POST -d '{ "id": "1", "withdrawalWitness":'$(cat data/withdrawal_witness_0x96da320fa4cd5e2f5d17f85e58e6e394f8c4b87cbb93e2af1fed50451094d8fd.json)', "prevWithdrawalProof":null }' -H "Content-Type: application/json" $WITHDRAWAL_PROVER_URL/proof/withdrawal | jq

# generate proof
curl -X POST -d '{ "id": "2", "withdrawalWitness":'$(cat data/withdrawal_witness_0x436d3f984fe2d267a6cf8bec2cd062473dd56cec08540115c430474c25f9be4e.json)', "prevWithdrawalProof":"'$(base64 --input data/prev_withdrawal_proof_0x436d3f984fe2d267a6cf8bec2cd062473dd56cec08540115c430474c25f9be4e.bin)'" }' -H "Content-Type: application/json" $WITHDRAWAL_PROVER_URL/proof/withdrawal | jq

# get proof
curl $WITHDRAWAL_PROVER_URL/proof/withdrawal/1 | jq

# get proofs
curl "$WITHDRAWAL_PROVER_URL/proofs/withdrawal?ids[]=1&ids[]=2" | jq
```

### Withdrawal Wrapper

```sh
# generate proof
curl -X POST -d '{ "id": "1", "withdrawalAggregator":"0x9Fa732B331a5455125c57f9db2E54E03c1CbbA5E", "withdrawalProof":"'$(base64 --input data/prev_withdrawal_proof_0x436d3f984fe2d267a6cf8bec2cd062473dd56cec08540115c430474c25f9be4e.bin)'" }' -H "Content-Type: application/json" $WITHDRAWAL_PROVER_URL/proof/wrapper | jq

# get proof
curl $WITHDRAWAL_PROVER_URL/proof/wrapper/1 | jq

# get proofs
curl "$WITHDRAWAL_PROVER_URL/proofs/wrapper?ids[]=1&ids[]=2" | jq
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