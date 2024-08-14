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

### Update (Synchronize Block)

```sh
# generate proof
curl -X POST -d '{ "balanceUpdateWitness":'$(cat data/balance_update_witness_0xb6958ba9425ec53e527c15d99420ec4e1af764aabed764a9435db4681e41b742.json)', "prevBalanceProof":null }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update | jq

# generate proof
curl -X POST -d '{ "balanceUpdateWitness":'$(cat data/balance_update_witness_0x5fdba28c55ab46d2acc036311e8835da80ae227c56c05aee645f5f2f1dda2443.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_proof_0x5fdba28c55ab46d2acc036311e8835da80ae227c56c05aee645f5f2f1dda2443.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x5fdba28c55ab46d2acc036311e8835da80ae227c56c05aee645f5f2f1dda2443.
curl http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update/0x5fdba28c55ab46d2acc036311e8835da80ae227c56c05aee645f5f2f1dda2443 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0xb6958ba9425ec53e527c15d99420ec4e1af764aabed764a9435db4681e41b742 or 0x5fdba28c55ab46d2acc036311e8835da80ae227c56c05aee645f5f2f1dda2443.
curl "http://localhost:8092/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update?blockHashes[]=0xb6958ba9425ec53e527c15d99420ec4e1af764aabed764a9435db4681e41b742&blockHashes[]=0x5fdba28c55ab46d2acc036311e8835da80ae227c56c05aee645f5f2f1dda2443" | jq
```

### Receive Deposit

```sh
# generate proof
curl -X POST -d '{ "receiveDepositWitness":'$(cat data/receive_deposit_witness_0.json)', "prevBalanceProof":"'$(base64 --input data/prev_receive_deposit_proof_0.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0
curl http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit/0 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and deposit index 0 or 1.
curl "http://localhost:8092/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/deposit?depositIndices[]=0&depositIndices[]=1" | jq
```

# Send Transaction

```sh
# generate proof
curl -X POST -d '{ "sendWitness":'$(cat data/send_witness_0x0913da753df6b870a294d2d277acc6e66ef5865db8a9dff6b27e32d9380144f4.json)', "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0x0913da753df6b870a294d2d277acc6e66ef5865db8a9dff6b27e32d9380144f4.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_for_send_proof_0x0913da753df6b870a294d2d277acc6e66ef5865db8a9dff6b27e32d9380144f4.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send | jq

# generate proof
curl -X POST -d '{ "sendWitness":'$(cat data/send_witness_0xe7d10c397020d2e484e57225943a64c24e88206f613a9f3e1956bebd61684080.json)', "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0xe7d10c397020d2e484e57225943a64c24e88206f613a9f3e1956bebd61684080.json)', "prevBalanceProof":"'$(base64 --input data/prev_balance_update_for_send_proof_0xe7d10c397020d2e484e57225943a64c24e88206f613a9f3e1956bebd61684080.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x6a04aacaa6f4492a806bf9cbf93bb3ac79975f06d5b92349ebef67f6f40c0cb9.
curl http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send/0xe7d10c397020d2e484e57225943a64c24e88206f613a9f3e1956bebd61684080 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x6a04aacaa6f4492a806bf9cbf93bb3ac79975f06d5b92349ebef67f6f40c0cb9 or 0xc9be81313526e0b29fe953f9b4feba4b05e2446d55fac9da92bda944c799333b.
curl "http://localhost:8092/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send?blockHashes[]=0x0913da753df6b870a294d2d277acc6e66ef5865db8a9dff6b27e32d9380144f4&blockHashes[]=0xe7d10c397020d2e484e57225943a64c24e88206f613a9f3e1956bebd61684080" | jq
```

### Receive Transfer (Synchronize Block)

```sh
# generate proof
curl -X POST -d '{ "receiveTransferWitness":'$(cat data/receive_transfer_witness_0x955146e44abdb771b50684e9c5af0746180ffbf62109df99310cba47ee41e72e.json)', "prevBalanceProof":"'$(base64 --input data/prev_receive_transfer_proof_0x955146e44abdb771b50684e9c5af0746180ffbf62109df99310cba47ee41e72e.bin)'" }' -H "Content-Type: application/json" http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transfer | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x955146e44abdb771b50684e9c5af0746180ffbf62109df99310cba47ee41e72e.
curl http://localhost:8092/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transfer/0x955146e44abdb771b50684e9c5af0746180ffbf62109df99310cba47ee41e72e | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and block hash 0x955146e44abdb771b50684e9c5af0746180ffbf62109df99310cba47ee41e72e or 0xc9be81313526e0b29fe953f9b4feba4b05e2446d55fac9da92bda944c799333b.
curl "http://localhost:8092/proofs/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transfer?privateCommitments[]=0x955146e44abdb771b50684e9c5af0746180ffbf62109df99310cba47ee41e72e&privateCommitments[]=0xc9be81313526e0b29fe953f9b4feba4b05e2446d55fac9da92bda944c799333b" | jq
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