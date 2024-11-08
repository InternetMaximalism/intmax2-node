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
curl -X POST -d '{ "requestId": "1", "balanceUpdateWitness":'$(cat data/update/single_update_witness_0x144b36dd812a1bbd4845c0e2bb7849fc9b398a957af375ccbcd9793b3d593985.json)', "prevBalanceProof":null }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x144b36dd812a1bbd4845c0e2bb7849fc9b398a957af375ccbcd9793b3d593985/update | jq

# generate proof (invalid)
curl -X POST -d '{ "requestId": "2", "balanceUpdateWitness":'$(cat data/update/update_witness_0x2257f6e44e3348c608f370172706c0f9ac8c60fa7754f77b1591d76b273b4a5d.json)', "prevBalanceProof":"'$(base64 --input data/update/balance_proof_send_from_spent_0x2257f6e44e3348c608f370172706c0f9ac8c60fa7754f77b1591d76b273b4a5d.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x2257f6e44e3348c608f370172706c0f9ac8c60fa7754f77b1591d76b273b4a5d/update | jq

# get the proof for public key 0x1d34947cdc0bbe768b1aa157ed75dc608b05012808e511840ee438e0e90d07f4 and ID 1.
curl $BALANCE_VALIDITY_PROVER_URL/proof/0x144b36dd812a1bbd4845c0e2bb7849fc9b398a957af375ccbcd9793b3d593985/update/1 | jq

# get the proof for public key 0x1ed9bc61cf8840c7e4a3fa12b330212ab3ab96eef6c07ead0a92c20ac5ed6242 and ID 1, 2.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0x2257f6e44e3348c608f370172706c0f9ac8c60fa7754f77b1591d76b273b4a5d/update?requestIds[]=1&requestIds[]=2" | jq
```

### Receive Deposit

```sh
# generate proof
curl -X POST -d '{ "requestId": "4", "receiveDepositWitness":'$(cat data/deposit/receive_deposit_witness_0x144b36dd812a1bbd4845c0e2bb7849fc9b398a957af375ccbcd9793b3d593985.json)', "prevBalanceProof":"'$(base64 --input data/deposit/balance_proof_receive_deposit_0x144b36dd812a1bbd4845c0e2bb7849fc9b398a957af375ccbcd9793b3d593985.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x144b36dd812a1bbd4845c0e2bb7849fc9b398a957af375ccbcd9793b3d593985/deposit | jq

# get the proof for public key 0x144b36dd812a1bbd4845c0e2bb7849fc9b398a957af375ccbcd9793b3d593985 and deposit index 4
curl $BALANCE_VALIDITY_PROVER_URL/proof/0x144b36dd812a1bbd4845c0e2bb7849fc9b398a957af375ccbcd9793b3d593985/deposit/4 | jq

# get the proof for public key 0x144b36dd812a1bbd4845c0e2bb7849fc9b398a957af375ccbcd9793b3d593985 and deposit index 4 or 5.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0x144b36dd812a1bbd4845c0e2bb7849fc9b398a957af375ccbcd9793b3d593985/deposit?requestIds[]=4&depositIndices[]=5" | jq
```

# Spent Transaction

```sh
# generate proof
curl -X POST -d '{ "requestId": "5", "spentWitness":'$(cat data/spend/spent_witness_0x23af9421582f7f19a52001f5c4f548da245dccd23da780c8b6f14bd285df1941.json)' }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/spend | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and ID 5.
curl $BALANCE_VALIDITY_PROVER_URL/proof/0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c/withdrawal/5 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and ID 5, 6.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c/withdrawal?requestIds[]=5&requestIds[]=6" | jq
```

# Send Transaction

```sh
# generate proof
curl -X POST -d '{ "requestId": "7", "txWitness": '$(cat data/send/tx_witness_0x2138b44f80a90601cbf10bce3e3e5fba760ee57153bff8dd546a299dff652fed.json)', "balanceUpdateWitness": '$(cat data/send/update_witness_0x2138b44f80a90601cbf10bce3e3e5fba760ee57153bff8dd546a299dff652fed.json)', "prevBalanceProof": "'$(base64 --input data/send/balance_proof_send_from_spent_0x2138b44f80a90601cbf10bce3e3e5fba760ee57153bff8dd546a299dff652fed.bin)'", "spentProof": "'$(base64 --input data/send/spent_proof_0x2138b44f80a90601cbf10bce3e3e5fba760ee57153bff8dd546a299dff652fed.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x2138b44f80a90601cbf10bce3e3e5fba760ee57153bff8dd546a299dff652fed/send | jq

# get the proof for public key 0x2138b44f80a90601cbf10bce3e3e5fba760ee57153bff8dd546a299dff652fed and ID 7.
curl $BALANCE_VALIDITY_PROVER_URL/proof/0x2138b44f80a90601cbf10bce3e3e5fba760ee57153bff8dd546a299dff652fed/send/7 | jq

# get the proof for public key 0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37 and ID 7 or 8.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0x2138b44f80a90601cbf10bce3e3e5fba760ee57153bff8dd546a299dff652fed/send?requestIds[]=7&requestIds[]=8" | jq
```

### Receive Transfer (Synchronize Block)

```sh
# generate proof
curl -X POST -d '{ "requestId": "9", "receiveTransferWitness":'$(cat data/transfer/receive_transfer_witness_0x238fd205205a0aee2d499f6bd7c3d97b875669ea9e06cb46a1ccf6513455f480.json)', "prevBalanceProof":"'$(base64 --input data/transfer/prev_balance_proof_receive_transfer_0x238fd205205a0aee2d499f6bd7c3d97b875669ea9e06cb46a1ccf6513455f480.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/0x238fd205205a0aee2d499f6bd7c3d97b875669ea9e06cb46a1ccf6513455f480/transfer | jq

# get the proof for public key 0x238fd205205a0aee2d499f6bd7c3d97b875669ea9e06cb46a1ccf6513455f480 and ID 9.
curl $BALANCE_VALIDITY_PROVER_URL/proof/0x238fd205205a0aee2d499f6bd7c3d97b875669ea9e06cb46a1ccf6513455f480/transfer/9 | jq

# get the proof for public key 0x238fd205205a0aee2d499f6bd7c3d97b875669ea9e06cb46a1ccf6513455f480 and ID 9, 10.
curl "$BALANCE_VALIDITY_PROVER_URL/proofs/0x238fd205205a0aee2d499f6bd7c3d97b875669ea9e06cb46a1ccf6513455f480/transfer?requestIds[]=9&requestIds[]=10" | jq
```

# Withdrawal Transaction

```sh
# generate proof
curl -X POST -d '{ "requestId": "10", "transferWitness":'$(cat data/withdrawal/withdrawal_witness_0x2138b44f80a90601cbf10bce3e3e5fba760ee57153bff8dd546a299dff652fed.json)', "balanceProof":"'$(base64 --input data/withdrawal/balance_proof_withdrawal_0x2138b44f80a90601cbf10bce3e3e5fba760ee57153bff8dd546a299dff652fed.bin)'" }' -H "Content-Type: application/json" $BALANCE_VALIDITY_PROVER_URL/proof/withdrawal | jq

curl $BALANCE_VALIDITY_PROVER_URL/proof/withdrawal/10 | jq

curl "$BALANCE_VALIDITY_PROVER_URL/proofs/withdrawal?requestIds[]=10&requestIds[]=11" | jq
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