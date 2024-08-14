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
# heath heck
curl http://localhost:8092/health | jq
```

### Receive Deposit

```sh
# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0
curl http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit/0 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0 or 1.
curl "http://localhost:8092/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit?depositIndices[]=0&depositIndices[]=1" | jq

# generate proof
curl -X POST -d '{ "receiveDepositWitness":'$(cat data/receive_deposit_witness_0.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_proof_0.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit | jq
```

### Update (Synchronize Block)

```sh
# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0xa33e362c4d3e8712cbc2a15cb7098b4b7d31d4698a1b71567040ddb4a0faca0f.
curl http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update/0xa33e362c4d3e8712cbc2a15cb7098b4b7d31d4698a1b71567040ddb4a0faca0f | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x5ed2018de7981aa199b5c31006007b41940796520c28bd06e7a64997d57e44d5 or 0xa33e362c4d3e8712cbc2a15cb7098b4b7d31d4698a1b71567040ddb4a0faca0f.
curl "http://localhost:8092/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update?blockHashes[]=0x5ed2018de7981aa199b5c31006007b41940796520c28bd06e7a64997d57e44d5&blockHashes[]=0xa33e362c4d3e8712cbc2a15cb7098b4b7d31d4698a1b71567040ddb4a0faca0f" | jq

# generate proof
curl -X POST -d '{ "balanceUpdateWitness":'$(cat data/balance_update_witness_0x5ed2018de7981aa199b5c31006007b41940796520c28bd06e7a64997d57e44d5.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_proof_0x5ed2018de7981aa199b5c31006007b41940796520c28bd06e7a64997d57e44d5.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update | jq

# generate proof
curl -X POST -d '{ "balanceUpdateWitness":'$(cat data/balance_update_witness_0xa33e362c4d3e8712cbc2a15cb7098b4b7d31d4698a1b71567040ddb4a0faca0f.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_proof_0xa33e362c4d3e8712cbc2a15cb7098b4b7d31d4698a1b71567040ddb4a0faca0f.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update | jq
```

### Receive Transfer (Synchronize Block)

```sh
# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x6a04aacaa6f4492a806bf9cbf93bb3ac79975f06d5b92349ebef67f6f40c0cb9.
curl http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transfer/0x6a04aacaa6f4492a806bf9cbf93bb3ac79975f06d5b92349ebef67f6f40c0cb9 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x6a04aacaa6f4492a806bf9cbf93bb3ac79975f06d5b92349ebef67f6f40c0cb9 or 0xc9be81313526e0b29fe953f9b4feba4b05e2446d55fac9da92bda944c799333b.
curl "http://localhost:8092/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transfer?blockHashes[]=0x6a04aacaa6f4492a806bf9cbf93bb3ac79975f06d5b92349ebef67f6f40c0cb9&blockHashes[]=0xc9be81313526e0b29fe953f9b4feba4b05e2446d55fac9da92bda944c799333b" | jq

# generate proof
curl -X POST -d '{ "receiveTransferWitness":'$(cat data/balance_receive_transfer_witness_0x6a04aacaa6f4492a806bf9cbf93bb3ac79975f06d5b92349ebef67f6f40c0cb9.json)', "prevBalanceProof":"'$(base64 --input data/prev_receive_transfer_proof_0x6a04aacaa6f4492a806bf9cbf93bb3ac79975f06d5b92349ebef67f6f40c0cb9.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transfer | jq
```

# Send Transaction

```sh
# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x6a04aacaa6f4492a806bf9cbf93bb3ac79975f06d5b92349ebef67f6f40c0cb9.
curl http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send/0xe7d10c397020d2e484e57225943a64c24e88206f613a9f3e1956bebd61684080 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x6a04aacaa6f4492a806bf9cbf93bb3ac79975f06d5b92349ebef67f6f40c0cb9 or 0xc9be81313526e0b29fe953f9b4feba4b05e2446d55fac9da92bda944c799333b.
curl "http://localhost:8092/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send?blockHashes[]=0x0913da753df6b870a294d2d277acc6e66ef5865db8a9dff6b27e32d9380144f4&blockHashes[]=0xe7d10c397020d2e484e57225943a64c24e88206f613a9f3e1956bebd61684080" | jq

# generate proof
curl -X POST -d '{ "sendWitness":'$(cat data/send_witness_0x0913da753df6b870a294d2d277acc6e66ef5865db8a9dff6b27e32d9380144f4.json)', "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0x0913da753df6b870a294d2d277acc6e66ef5865db8a9dff6b27e32d9380144f4.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_for_send_proof_0x0913da753df6b870a294d2d277acc6e66ef5865db8a9dff6b27e32d9380144f4.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send | jq

# generate proof
curl -X POST -d '{ "sendWitness":'$(cat data/send_witness_0xe7d10c397020d2e484e57225943a64c24e88206f613a9f3e1956bebd61684080.json)', "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0xe7d10c397020d2e484e57225943a64c24e88206f613a9f3e1956bebd61684080.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_for_send_proof_0xe7d10c397020d2e484e57225943a64c24e88206f613a9f3e1956bebd61684080.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send | jq
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