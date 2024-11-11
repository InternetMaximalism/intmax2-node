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
curl -X POST -d '{ "id": "1", "singleWithdrawalProof": "'$(base64 --input data/single_withdrawal_proof.bin)'", "prevWithdrawalProof": null }' -H "Content-Type: application/json" $WITHDRAWAL_PROVER_URL/proof/withdrawal | jq

# generate proof
curl -X POST -d '{ "id": "2", "singleWithdrawalProof": "'$(base64 --input data/single_withdrawal_proof.bin)'", "prevWithdrawalProof": '$(cat data/prev_withdrawal_proof.json)' }' -H "Content-Type: application/json" $WITHDRAWAL_PROVER_URL/proof/withdrawal | jq

# get proof
curl $WITHDRAWAL_PROVER_URL/proof/withdrawal/1 | jq .data.proof

# get proofs
curl "$WITHDRAWAL_PROVER_URL/proofs/withdrawal?ids[]=1&ids[]=2" | jq
```

### Withdrawal Wrapper

```sh
# generate proof
curl -X POST -d '{ "id": "1", "withdrawalAggregator": "0x420a5b76e11e80d97c7eb3a0b16ac7b70672b8c2", "withdrawalProof": "'$(base64 --input data/withdrawal_proof.bin)'" }' -H "Content-Type: application/json" $WITHDRAWAL_PROVER_URL/proof/wrapper | jq

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