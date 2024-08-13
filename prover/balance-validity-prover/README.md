# balance-validity-prover

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

### Deposit

```sh
# heath heck
curl http://localhost:8092/health | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0
curl http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit/0 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0 or 1.
curl "http://localhost:8092/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit?depositIndices[]=0&depositIndices[]=1" | jq

# generate proof
curl -X POST -d '{ "receiveDepositWitness":'$(cat data/receive_deposit_witness_0.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_proof_0.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit | jq
```

### Update (Synchronize Block)

```sh
# heath heck
curl http://localhost:8092/health | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0
curl http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update/0xa33e362c4d3e8712cbc2a15cb7098b4b7d31d4698a1b71567040ddb4a0faca0f | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0 or 1.
curl "http://localhost:8092/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update?blockHashes[]=0x5ed2018de7981aa199b5c31006007b41940796520c28bd06e7a64997d57e44d5&blockHashes[]=0xa33e362c4d3e8712cbc2a15cb7098b4b7d31d4698a1b71567040ddb4a0faca0f" | jq

# generate proof
curl -X POST -d '{ "balanceUpdateWitness":'$(cat data/balance_update_witness_0x5ed2018de7981aa199b5c31006007b41940796520c28bd06e7a64997d57e44d5.json)', "prevBalanceUpdateProof":"'$(base64 --input data/prev_balance_update_proof_0x5ed2018de7981aa199b5c31006007b41940796520c28bd06e7a64997d57e44d5.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update | jq

# generate proof
curl -X POST -d '{ "balanceUpdateWitness":'$(cat data/balance_update_witness_0xa33e362c4d3e8712cbc2a15cb7098b4b7d31d4698a1b71567040ddb4a0faca0f.json)', "prevBalanceUpdateProof":"'$(base64 --input data/prev_balance_update_proof_0xa33e362c4d3e8712cbc2a15cb7098b4b7d31d4698a1b71567040ddb4a0faca0f.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update | jq
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