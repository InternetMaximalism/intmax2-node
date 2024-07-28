# withdraw-aggregator-prover

## Development

```sh
# env
cp .env.example .env

# run app
cargo run .
```

## APIs

```sh
# heath heck
curl http://localhost:8080/health | jq

# get proof
curl http://localhost:8080/proof/1 | jq

# get proofs
curl "http://localhost:8080/proofs?ids[]=1&ids[]=2" | jq

# generate proof
curl -X POST -d '{"id":"1"}' -H "Content-Type: application/json" http://localhost:8080/proof | jq
curl -X POST -d '{"id":"2"}' -H "Content-Type: application/json" http://localhost:8080/proof | jq
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