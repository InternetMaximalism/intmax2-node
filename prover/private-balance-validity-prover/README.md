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

### Receive Deposit

```sh
# generate proof
curl -X POST -d '{ "receiveDepositWitness":'$(cat data/receive_deposit_witness_0xfe016b28057ea074001cc9ce16e323d47e7608791a4d102ce6d8283683decb63.json)', "prevBalanceProof":"'$(base64 --input data/prev_receive_deposit_proof_0xfe016b28057ea074001cc9ce16e323d47e7608791a4d102ce6d8283683decb63.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transition/deposit | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0
curl $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transition/deposit/0xfe016b28057ea074001cc9ce16e323d47e7608791a4d102ce6d8283683decb63 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0 or 1.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transition/deposit?ids[]=0xfe016b28057ea074001cc9ce16e323d47e7608791a4d102ce6d8283683decb63&ids[]=0xc9be81313526e0b29fe953f9b4feba4b05e2446d55fac9da92bda944c799333b" | jq
```

# Send Transaction

```sh
# generate proof
curl -X POST -d '{ "sendWitness":'$(cat data/send_witness_0xc2c45b592d56f14cb8574e4a90392bf50b81ad73dc31a9203c240cad74d7a491.json)', "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0xc2c45b592d56f14cb8574e4a90392bf50b81ad73dc31a9203c240cad74d7a491.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_for_send_proof_0xc2c45b592d56f14cb8574e4a90392bf50b81ad73dc31a9203c240cad74d7a491.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transition/send | jq

# generate proof
curl -X POST -d '{ "sendWitness":'$(cat data/send_witness_0x395a3add5dea1151b6ffaf3b532d9a4ae0337d3c8f33d80140c1daf6d7b2084e.json)', "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0x395a3add5dea1151b6ffaf3b532d9a4ae0337d3c8f33d80140c1daf6d7b2084e.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_for_send_proof_0x395a3add5dea1151b6ffaf3b532d9a4ae0337d3c8f33d80140c1daf6d7b2084e.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transition/send | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x395a3add5dea1151b6ffaf3b532d9a4ae0337d3c8f33d80140c1daf6d7b2084e.
curl $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transition/send/0x395a3add5dea1151b6ffaf3b532d9a4ae0337d3c8f33d80140c1daf6d7b2084e | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0xc2c45b592d56f14cb8574e4a90392bf50b81ad73dc31a9203c240cad74d7a491 or 0x395a3add5dea1151b6ffaf3b532d9a4ae0337d3c8f33d80140c1daf6d7b2084e.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transition/send?ids[]=0xc2c45b592d56f14cb8574e4a90392bf50b81ad73dc31a9203c240cad74d7a491&ids[]=0x395a3add5dea1151b6ffaf3b532d9a4ae0337d3c8f33d80140c1daf6d7b2084e" | jq
```

### Receive Transfer (Synchronize Block)

```sh
# generate proof
curl -X POST -d '{ "receiveTransferWitness":'$(cat data/receive_transfer_witness_0xfe016b28057ea074001cc9ce16e323d47e7608791a4d102ce6d8283683decb63.json)', "prevBalanceProof":"'$(base64 --input data/prev_receive_transfer_proof_0xfe016b28057ea074001cc9ce16e323d47e7608791a4d102ce6d8283683decb63.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transition/transfer | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0xfe016b28057ea074001cc9ce16e323d47e7608791a4d102ce6d8283683decb63.
curl $BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transition/transfer/0xfe016b28057ea074001cc9ce16e323d47e7608791a4d102ce6d8283683decb63 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0xfe016b28057ea074001cc9ce16e323d47e7608791a4d102ce6d8283683decb63 or 0xc9be81313526e0b29fe953f9b4feba4b05e2446d55fac9da92bda944c799333b.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transition/transfer?ids[]=0xfe016b28057ea074001cc9ce16e323d47e7608791a4d102ce6d8283683decb63&ids[]=0xc9be81313526e0b29fe953f9b4feba4b05e2446d55fac9da92bda944c799333b" | jq
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