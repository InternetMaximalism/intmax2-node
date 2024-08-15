# block-validity-prover

## Development

```sh
# env
cp .env.example .env

# setup nightly
rustup override set nightly

# run app
cargo run -r --features dummy_proof
```

## APIs

```sh
# set API URL
BLOCK_VALIDITY_PROVER_URL="http://localhost:8080"

# heath heck
curl $BLOCK_VALIDITY_PROVER_URL/health | jq

# generate proof
curl -X POST -d '{"blockHash":"0x01", "validityWitness":'$(cat data/validity_witness_1.json)', "prevValidityProof":null }' -H "Content-Type: application/json" $BLOCK_VALIDITY_PROVER_URL/proof | jq

# generate proof
curl -X POST -d '{"blockHash":"0x02", "validityWitness":'$(cat data/validity_witness_2.json)', "prevValidityProof":"'$(base64 --input data/prev_validity_proof_2.bin)'" }' -H "Content-Type: application/json" $BLOCK_VALIDITY_PROVER_URL/proof | jq

# generate proof
curl -X POST -d '{"blockHash":"0x03", "validityWitness":'$(cat data/validity_witness_3.json)', "prevValidityProof":"'$(base64 --input data/prev_validity_proof_3.bin)'" }' -H "Content-Type: application/json" $BLOCK_VALIDITY_PROVER_URL/proof | jq

# get proof
curl $BLOCK_VALIDITY_PROVER_URL/proof/0x01 | jq

# get proofs
curl "$BLOCK_VALIDITY_PROVER_URL/proofs?blockHashes[]=0x02&blockHashes[]=0x03" | jq
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