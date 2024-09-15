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

### Health Check

```sh
BALANCE_VALIDITY_PROVER_URL="http://localhost:8080"

# heath heck
curl $BALANCE_VALIDITY_PROVER_URL/health | jq
```

### Update (Synchronize Block)

```sh
# generate proof
curl -X POST -d '{ "requestId": "1", "balanceUpdateWitness":'$(cat data/balance_update_witness_0xb0f9cbdf7b1f89cad6d6657520505a117ac69b834d502ca9b1ecfb3f1bfa5556.json)', "prevBalanceProof":null }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update | jq

# generate proof
curl -X POST -d '{ "requestId": "2", "balanceUpdateWitness":'$(cat data/balance_update_witness_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_proof_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update | jq

# generate proof (XXX: invalid)
curl -X POST -d '{ "requestId": "3", "balanceUpdateWitness":'$(cat data/balance_update_witness_0x2fc9d0cc9b9a135ea38a2fa0260406dcd4d9e65678c102d7c439e2a50401d217.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_proof_0x2fc9d0cc9b9a135ea38a2fa0260406dcd4d9e65678c102d7c439e2a50401d217.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0xb0f9cbdf7b1f89cad6d6657520505a117ac69b834d502ca9b1ecfb3f1bfa5556.
curl $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update/0xb0f9cbdf7b1f89cad6d6657520505a117ac69b834d502ca9b1ecfb3f1bfa5556 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c or 0xb0f9cbdf7b1f89cad6d6657520505a117ac69b834d502ca9b1ecfb3f1bfa5556.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update?blockHashes[]=0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c&blockHashes[]=0xb0f9cbdf7b1f89cad6d6657520505a117ac69b834d502ca9b1ecfb3f1bfa5556" | jq
```

### Receive Deposit

```sh
# generate proof
curl -X POST -d '{ "requestId": "1", "receiveDepositWitness":'$(cat data/receive_deposit_witness_0.json)', "prevBalanceProof":"'$(base64 --input data/prev_receive_deposit_proof_0.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0
curl $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit/0 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0 or 1.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit?depositIndices[]=0&depositIndices[]=1" | jq
```

# Spent Transaction

```sh
# generate proof
curl -X POST -d '{ "requestId": "1", "sendWitness":'$(cat data/send_witness_0x2fc9d0cc9b9a135ea38a2fa0260406dcd4d9e65678c102d7c439e2a50401d217.json)', "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0x2fc9d0cc9b9a135ea38a2fa0260406dcd4d9e65678c102d7c439e2a50401d217.json)' }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/spent | jq

# generate proof
curl -X POST -d '{ "requestId": "2", "sendWitness":'$(cat data/send_witness_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.json)', "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.json)' }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/send | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.
curl $BALANCE_VALIDITY_PROVER_URL/proof/spent/0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x2fc9d0cc9b9a135ea38a2fa0260406dcd4d9e65678c102d7c439e2a50401d217 or 0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/spent?requestId[]=0x2fc9d0cc9b9a135ea38a2fa0260406dcd4d9e65678c102d7c439e2a50401d217&requestId[]=0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c" | jq
```

# Send Transaction

```sh
# generate proof
curl -X POST -d '{ "requestId": "1", "sendWitness":'$(cat data/send_witness_0x2fc9d0cc9b9a135ea38a2fa0260406dcd4d9e65678c102d7c439e2a50401d217.json)', "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0x2fc9d0cc9b9a135ea38a2fa0260406dcd4d9e65678c102d7c439e2a50401d217.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_for_send_proof_0x2fc9d0cc9b9a135ea38a2fa0260406dcd4d9e65678c102d7c439e2a50401d217.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send | jq

# generate proof
curl -X POST -d '{ "requestId": "2", "sendWitness":'$(cat data/send_witness_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.json)', "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_for_send_proof_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.
curl $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send/0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x2fc9d0cc9b9a135ea38a2fa0260406dcd4d9e65678c102d7c439e2a50401d217 or 0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send?blockHashes[]=0x2fc9d0cc9b9a135ea38a2fa0260406dcd4d9e65678c102d7c439e2a50401d217&blockHashes[]=0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c" | jq
```

### Receive Transfer (Synchronize Block)

```sh
# generate proof
curl -X POST -d '{ "requestId": "1", "receiveTransferWitness":'$(cat data/receive_transfer_witness_0x7a00b7dbf1994ff9fb05a5897b7dc459dd9167ee7a4ad049b9850cbaf286bbee.json)', "prevBalanceProof":"'$(base64 --input data/prev_receive_transfer_proof_0x7a00b7dbf1994ff9fb05a5897b7dc459dd9167ee7a4ad049b9850cbaf286bbee.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transfer | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x7a00b7dbf1994ff9fb05a5897b7dc459dd9167ee7a4ad049b9850cbaf286bbee.
curl $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transfer/0x7a00b7dbf1994ff9fb05a5897b7dc459dd9167ee7a4ad049b9850cbaf286bbee | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x7a00b7dbf1994ff9fb05a5897b7dc459dd9167ee7a4ad049b9850cbaf286bbee or 0xc9be81313526e0b29fe953f9b4feba4b05e2446d55fac9da92bda944c799333b.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transfer?privateCommitments[]=0x7a00b7dbf1994ff9fb05a5897b7dc459dd9167ee7a4ad049b9850cbaf286bbee&privateCommitments[]=0xc9be81313526e0b29fe953f9b4feba4b05e2446d55fac9da92bda944c799333b" | jq
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