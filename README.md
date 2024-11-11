# INTMAX2 Node

## Requirements
* Docker v24.0.2+;
* Docker Compose v2;
* Go v1.21+;

## Quick Start
To run and manage microservice, you can use Makefile, which contains already created commands to run, build and remove containers, also for local compilation.
* Local build: `make build-up`
* Local start:
```
(before, need copies `.env.example` as `.env`)
# make tools
# make gen
# make start-build-up
# SWAGGER_USE=true GIT_USE=true CMD_RUN="migrate --action=up" make run
# SWAGGER_USE=true GIT_USE=true CMD_RUN="run" make run
```

## Makefile command list
* `make tools`
* `make gen`
* `make build-up`
* `make up`
* `make lint`
* `make down`

## Swagger build customization with Dockerfile
```
SWAGGER_HOST_URL=127.0.0.1:8780 \
SWAGGER_BASE_PATH="\/" \
SWAGGER_USE=true \
make build-up
``` 

## Service connections

### Connections
* (node) Serving gRPC server on 0.0.0.0:10000
* (node) Serving HTTP on http://0.0.0.0:80
* (node) Serving status on http://0.0.0.0:80/status
* (node) Serving health on http://0.0.0.0:80/health
* (node) Serving prometheus metric on http://0.0.0.0:80/prometheus
* (node) Serving OpenAPI Documentation on http://0.0.0.0:80/swagger/
* (node) Serving JSON OpenAPI Documentation on http://0.0.0.0:80/node/apidocs.swagger.json

## Configuration
Available Commands:
### Command `./intmax2-node --help`
```
# ./intmax2-node --help
Usage:
  app [flags]
  app [command]

Available Commands:
  balance                     Manage balance
  block_builder               Manage block builder
  completion                  Generate the autocompletion script for the specified shell
  deposit                     Manage deposit
  ethereum_private_key_wallet Generate Ethereum and IntMax wallets from Ethereum private key
  generate_account            Generate new Ethereum and IntMax accounts
  help                        Help about any command
  intmax_private_key_wallet   Generate IntMax wallet from IntMax private key
  messenger                   Manage messenger
  migrate                     Execute migration
  mnemonic_account            Generate Ethereum and IntMax accounts from mnemonic
  run                         run command
  store-vault-server          run store valut server command
  tx                          Manage transaction
  withdrawal                  Manage withdrawal
  withdrawal-server           run withdrawal server command

Flags:
  -h, --help   help for app

Use "app [command] --help" for more information about a command.

```
### Command `./intmax2-node run --help`
```
# ./intmax2-node run --help
run command

Usage:
app run [flags]

Flags:
  -h, --help   help for run
```
### Command `./intmax2-node migrate --help`
```
# ./intmax2-node migrate --help
Execute migrations stored at binary
Actions:
up - migrate all steps Up
down - migrate all steps Down
number - amount of steps to migrate (if > 0 - migrate number steps up, if < 0 migrate number steps down)

Usage:
  app migrate --action "<up|down|1|-1>" [flags]

Flags:
      --action string   action flag. use as --action "<up|down|1|-1>"
  -h, --help            help for migrate
```
### Command `./intmax2-node generate_account --help`
```
# ./intmax2-node generate_account --help
Generate new Ethereum and IntMax accounts

Usage:
  app generate_account [flags]

Flags:
      --derivation_path string     derivation_path flag. use as --derivation_path "m/44'/60'/0'/0/" (default "m/44'/60'/0'/0/")
  -h, --help                       help for generate_account
      --key_number string          key_number flag. use as --key_number "0" (0 - parent account, 1...n - child accounts numbers)
      --mnemonic_password string   mnemonic_password flag. use as --mnemonic_password "pass"
```
### Command `./intmax2-node mnemonic_account --help`
```
# ./intmax2-node mnemonic_account --help
Generate Ethereum and IntMax accounts from mnemonic

Usage:
  app mnemonic_account [flags]

Flags:
      --derivation_path string     derivation_path flag. use as --derivation_path "m/44'/60'/0'/0/" (default "m/44'/60'/0'/0/")
  -h, --help                       help for mnemonic_account
      --key_number string          key_number flag. use as --key_number "0" (0 - parent account, 1...n - child accounts numbers)
      --mnemonic string            mnemonic flag. use as --mnemonic "mnemonic1 mnemonic2 ... mnemonic24"
      --mnemonic_password string   mnemonic_password flag. use as --mnemonic_password "pass"
```
### Command `./intmax2-node ethereum_private_key_wallet --help`
```
# ./intmax2-node ethereum_private_key_wallet --help
Generate Ethereum and IntMax wallets from Ethereum private key

Usage:
  app private_key_wallet [flags]

Flags:
  -h, --help                 help for private_key_wallet
      --private_key string   private_key flag. use as --private_key "__PRIVATE_KEY_IN_HEX_WITHOUT_0x__"
```
### Command `./intmax2-node block_builder --help`
```
# ./intmax2-node block_builder --help
Manage block builder

Usage:
  app block_builder [command]

Available Commands:
  info        Returns the block builder info
  stop        Stop block builder
  unstake     Unstake block builder

Flags:
  -h, --help   help for block_builder

Use "app block_builder [command] --help" for more information about a command.
```
### Command `./intmax2-node block_builder info --help`
```
# ./intmax2-node block_builder info --help
Returns the block builder info

Usage:
  app block_builder info [flags]

Flags:
  -h, --help   help for info
```
### Command `./intmax2-node block_builder stop --help`
```
# ./intmax2-node block_builder stop --help
Stop block builder

Usage:
  app block_builder stop [flags]

Flags:
  -h, --help   help for stop
```
### Command `./intmax2-node block_builder unstake --help`
```
# ./intmax2-node block_builder unstake --help
Unstake block builder

Usage:
  app block_builder unstake [flags]

Flags:
  -h, --help   help for unstake
```
### Command `./intmax2-node deposit --help`
```
# ./intmax2-node deposit --help
Manage deposit

Usage:
  app deposit [command]

Available Commands:
  analyzer    Run deposit analyzer service
  relayer     Run deposit relayer service

Flags:
  -h, --help   help for deposit

Use "app deposit [command] --help" for more information about a command.
```
### Command `./intmax2-node deposit analyzer --help`
```
# ./intmax2-node deposit analyzer --help
Run deposit analyzer service

Usage:
  app deposit analyzer [flags]

Flags:
  -h, --help   help for analyzer
```
### Command `./intmax2-node store-valut-server run --help`
```
# ./intmax2-node store-valut-server run --help
Run store vault server

Usage:
  app store-valut-server run [flags]

Flags:
  -h, --help help for server
```
### Command `./intmax2-node withdrawal-server run --help`
```
# ./intmax2-node withdrawal-server run --help
Run withdrawal server

Usage:
  app withdrawal-server run [flags]

Flags:
  -h, --help help for server
```
### Command `./intmax2-node withdrawal aggregator --help`
```
# ./intmax2-node withdrawal aggregator --help
Run withdrawal aggregator service

Usage:
  app withdrawal aggregator [flags]

Flags:
  -h, --help help for aggregator
```
### Command `./intmax2-node messenger withdrawal-relayer --help`
```
# ./intmax2-node withdrawal withdrawal-relayer --help
Run messenger withdrawal-relayer service

Usage:
  app messenger withdrawal-relayer [flags]

Flags:
  -h, --help help for withdrawal-relayer
```
### Command `./intmax2-node messenger withdrawal-relayer-mock --help`
```
# ./intmax2-node withdrawal withdrawal-relayer-mock --help
Run messenger withdrawal-relayer-mock service

Usage:
  app messenger withdrawal-relayer-mock [flags]

Flags:
  -h, --help help for withdrawal-relayer-mock
```
### Command `./intmax2-node balance --help`
```
# ./intmax2-node balance --help
Manage balance

Usage:
  app balance [command]

Available Commands:
  get         Manage of Get balance of specified INTMAX account

Flags:
  -h, --help   help for balance

Use "app balance [command] --help" for more information about a command.
```
### Command `./intmax2-node balance get --help`
```
# ./intmax2-node balance get --help
Manage of Get balance of specified INTMAX account

Usage:
  app balance get [command]

Available Commands:
  erc1155     Get balance by token "erc1155" of specified INTMAX account
  erc20       Get balance by token "erc20" of specified INTMAX account
  erc721      Get balance by token "erc721" of specified INTMAX account
  eth         Get balance by token "eth" of specified INTMAX account

Flags:
  -h, --help   help for get

Use "app balance get [command] --help" for more information about a command.
```
### Command `./intmax2-node balance get erc1155 --help`
```
# ./intmax2-node balance get erc1155 --help
Get balance by token "erc1155" of specified INTMAX account

Usage:
  app balance get erc1155 [flags]

Flags:
  -h, --help                 help for erc1155
      --private-key string   specify user address. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
```
### Command `./intmax2-node balance get erc20 --help`
```
# ./intmax2-node balance get erc20 --help
Get balance by token "erc20" of specified INTMAX account

Usage:
  app balance get erc20 [flags]

Flags:
  -h, --help                 help for erc20
      --private-key string   specify user address. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"

Example:
  ./intmax2-node balance get erc20 0x0000000000000000000000000000000000000001 --private-key 0x030644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd3
```
### Command `./intmax2-node balance get erc721 --help`
```
# ./intmax2-node balance get erc721 --help
Get balance by token "erc721" of specified INTMAX account

Usage:
  app balance get erc721 [flags]

Flags:
  -h, --help                 help for erc721
      --private-key string   specify user address. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
```
### Command `./intmax2-node balance get eth --help`
```
# ./intmax2-node balance get eth --help
Get balance by token "eth" of specified INTMAX account

Usage:
  app balance get eth [flags]

Flags:
  -h, --help                 help for eth
      --private-key string   specify user address. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"

Example:
  ./intmax2-node balance get eth --private-key 0x030644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd3
```
### Command `./intmax2-node tx deposit --help`
```
# ./intmax2-node tx deposit --help
Manage deposit transaction

Usage:
  app tx deposit [command]

Available Commands:
  erc1155     Send deposit transaction by token "erc1155"
  erc20       Send deposit transaction by token "erc20"
  erc721      Send deposit transaction by token "erc721"
  eth         Send deposit transaction by token "eth"
  info        Manage deposit by hash
  list        Manage deposit list

Flags:
  -h, --help   help for deposit
```
### Command `./intmax2-node tx deposit eth --help`
```
# ./intmax2-node tx deposit eth --help
Send deposit transaction by token "eth"

Usage:
  app tx deposit eth [flags]

Flags:
      --amount string        specify amount without decimals. use as --amount "10"
  -h, --help                 help for eth
      --private-key string   specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
      --recipient string     specify recipient INTMAX address. use as --recipient "0x0000000000000000000000000000000000000000000000000000000000000000"

Example:
  ./intmax2-node tx deposit eth --amount 10000 --recipient 0x06a7b64af8f414bcbeef455b1da5208c9b592b83ee6599824caa6d2ee9141a76 --private-key 0x0000000000000000000000000000000000000000000000000000000000000002
```
### Command `./intmax2-node tx deposit erc20 --help`
```
# ./intmax2-node tx deposit erc20 --help
Send deposit transaction by token "erc20"

Usage:
  app tx deposit erc20 [flags]

Flags:
      --amount string        specify amount without decimals. use as --amount "10"
  -h, --help                 help for erc20
      --private-key string   specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
      --recipient string     specify recipient INTMAX address. use as --recipient "0x0000000000000000000000000000000000000000000000000000000000000000"

Example:
  ./intmax2-node tx deposit erc20 0x0000000000000000000000000000000000000001 --amount 10000 --recipient 0x06a7b64af8f414bcbeef455b1da5208c9b592b83ee6599824caa6d2ee9141a76 --private-key 0x0000000000000000000000000000000000000000000000000000000000000002
```
### Command `./intmax2-node tx deposit erc721 --help`
```
# ./intmax2-node tx deposit erc721 --help
Send deposit transaction by token "erc721"

Usage:
  app tx deposit erc721 [flags]

Flags:
      --amount string        specify amount without decimals. use as --amount "10"
  -h, --help                 help for erc721
      --private-key string   specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
      --recipient string     specify recipient INTMAX address. use as --recipient "0x0000000000000000000000000000000000000000000000000000000000000000"

Example:
  ./intmax2-node tx deposit erc721 0x0000000000000000000000000000000000000001 7 --amount 10000 --recipient 0x06a7b64af8f414bcbeef455b1da5208c9b592b83ee6599824caa6d2ee9141a76 --private-key 0x0000000000000000000000000000000000000000000000000000000000000002
```
### Command `./intmax2-node tx deposit erc1155 --help`
```
# ./intmax2-node tx deposit erc1155 --help
Send deposit transaction by token "erc1155"

Usage:
  app tx deposit erc1155 [flags]

Flags:
      --amount string        specify amount without decimals. use as --amount "10"
  -h, --help                 help for erc1155
      --private-key string   specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
      --recipient string     specify recipient INTMAX address. use as --recipient "0x0000000000000000000000000000000000000000000000000000000000000000"
```
### Command `./intmax2-node tx deposit list --help`
```
# ./intmax2-node tx deposit list --help
Manage deposit list

Usage:
  app tx deposit list [command]

Available Commands:
  incoming    Get deposit list (incoming)
  outgoing    Get deposit list (outgoing); coming soon

Flags:
  -h, --help   help for list

Use "app tx deposit list [command] --help" for more information about a command.
```
### Command `./intmax2-node tx deposit list incoming --help`
```
# ./intmax2-node tx deposit list incoming --help
Get deposit list (incoming)

Usage:
  app tx deposit list incoming [flags]

Flags:
      --filterCondition string                specify the filter condition. use as --filterCondition "is" (support values: "lessThan", "lessThanOrEqualTo", "is", "greaterThanOrEqualTo", "greaterThan")
      --filterName string                     specify the filter name. use as --filterName "block_number" (support value: "block_number")
      --filterValue string                    specify the value of filter. use as --filterValue "1"
  -h, --help                                  help for incoming
      --paginationCursorBlockNumber string    specify the BlockNumber cursor. use as --paginationCursorBlockNumber "1" (more then "0")
      --paginationCursorSortingValue string   specify the SortingValue cursor. use as --paginationCursorSortingValue "1" (more then "0")
      --paginationDirection string            specify the direction pagination. use as --paginationDirection "next" (support values: "next", "prev") (default "next")
      --paginationLimit string                specify the limit for pagination without decimals. use as --paginationLimit "100" (default "100")
      --private-key string                    specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
      --sorting string                        specify the sorting. use as --sorting "desc" (support values: "asc", "desc") (default "desc")
```
### Command `./intmax2-node tx deposit info --help`
```
# ./intmax2-node tx deposit info --help
Manage deposit by hash

Usage:
  app tx deposit info [command]

Available Commands:
  incoming    Get deposit by hash (incoming)
  outgoing    Get deposit by hash (outgoing); coming soon

Flags:
  -h, --help   help for info

Use "app tx deposit info [command] --help" for more information about a command.
```
### Command `./intmax2-node tx deposit info incoming --help`
```
# ./intmax2-node tx deposit info incoming --help
Get deposit by hash (incoming)

Usage:
  app tx deposit info incoming [DepositHash] [flags]

Flags:
  -h, --help                 help for incoming
      --private-key string   specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
```
### Command `./intmax2-node tx transfer --help`
```
# ./intmax2-node tx transfer --help
Manage transfer transaction

Usage:
  app tx transfer [command]

Available Commands:
  erc1155     Send transfer transaction by token "erc1155"
  erc20       Send transfer transaction by token "erc20"
  erc721      Send transfer transaction by token "erc721"
  eth         Send transfer transaction by token "eth"
  info        Get transaction by hash
  list        Get transactions list

Flags:
  -h, --help   help for transfer
```
### Command `./intmax2-node tx transfer eth --help`
```
Send transfer transaction by token "eth"

Usage:
  app tx transfer eth [flags]

Flags:
      --amount string        specify amount without decimals. use as --amount "10"
  -h, --help                 help for eth
      --private-key string   specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
      --recipient string     specify recipient INTMAX address. use as --recipient "0x0000000000000000000000000000000000000000000000000000000000000000"

Example:
  ./intmax2-node tx transfer eth --amount 10 --recipient 0x06a7b64af8f414bcbeef455b1da5208c9b592b83ee6599824caa6d2ee9141a76 --private-key 0x0000000000000000000000000000000000000000000000000000000000000002
```
### Command `./intmax2-node tx transfer erc20 --help`
```
Send transfer transaction by token "erc20"

Usage:
  app tx transfer erc20 [flags]

Flags:
      --amount string        specify amount without decimals. use as --amount "10"
  -h, --help                 help for erc20
      --private-key string   specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
      --recipient string     specify recipient INTMAX address. use as --recipient "0x0000000000000000000000000000000000000000000000000000000000000000"

Example:
  ./intmax2-node tx transfer erc20 0x0000000000000000000000000000000000000001 --amount 10 --recipient 0x06a7b64af8f414bcbeef455b1da5208c9b592b83ee6599824caa6d2ee9141a76 --private-key 0x0000000000000000000000000000000000000000000000000000000000000002
```
### Command `./intmax2-node tx transfer erc721 --help`
```
Send transfer transaction by token "erc721"

Usage:
  app tx transfer erc721 [flags]

Flags:
      --amount string        specify amount without decimals. use as --amount "10"
  -h, --help                 help for erc721
      --private-key string   specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
      --recipient string     specify recipient INTMAX address. use as --recipient "0x0000000000000000000000000000000000000000000000000000000000000000"

Example:
  ./intmax2-node tx transfer erc721 0x0000000000000000000000000000000000000001 7 --amount 10 --recipient 0x06a7b64af8f414bcbeef455b1da5208c9b592b83ee6599824caa6d2ee9141a76 --private-key 0x0000000000000000000000000000000000000000000000000000000000000002
```
### Command `./intmax2-node tx transfer erc1155 --help`
```
Send transfer transaction by token "erc1155"

Usage:
  app tx transfer erc1155 [flags]

Flags:
      --amount string        specify amount without decimals. use as --amount "10"
  -h, --help                 help for erc1155
      --private-key string   specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
      --recipient string     specify recipient INTMAX address. use as --recipient "0x0000000000000000000000000000000000000000000000000000000000000000"
```
### Command `./intmax2-node tx transfer list --help`
```
Get transactions list

Usage:
  app tx transfer list [flags]

Flags:
      --filterCondition string                specify the filter condition. use as --filterCondition "is" (support values: "lessThan", "lessThanOrEqualTo", "is", "greaterThanOrEqualTo", "greaterThan")
      --filterName string                     specify the filter name. use as --filterName "block_number" (support value: "block_number")
      --filterValue string                    specify the value of filter. use as --filterValue "1"
  -h, --help                                  help for list
      --paginationCursorBlockNumber string    specify the BlockNumber cursor. use as --paginationCursorBlockNumber "1" (more then "0")
      --paginationCursorSortingValue string   specify the SortingValue cursor. use as --paginationCursorSortingValue "1" (more then "0")
      --paginationDirection string            specify the direction pagination. use as --paginationDirection "next" (support values: "next", "prev") (default "next")
      --paginationLimit string                specify the limit for pagination without decimals. use as --paginationLimit "100" (default "100")
      --private-key string                    specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
      --sorting string                        specify the sorting. use as --sorting "desc" (support values: "asc", "desc") (default "desc")
```
### Command `./intmax2-node tx transfer info --help`
```
Get transaction by hash

Usage:
  app tx transfer info [TxHash] [flags]

Flags:
  -h, --help                 help for info
      --private-key string   specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
```
### Command `./intmax2-node tx withdrawal --help`
```
# ./intmax2-node tx withdrawal --help
Send withdrawal transaction

Usage:
  app tx withdrawal [flags]

Flags:
      --amount string        specify amount without decimals. use as --amount "10"
  -h, --help                 help for withdrawal
      --private-key string   specify user's private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
      --recipient string     specify recipient Ethereum address. use as --recipient "0x0000000000000000000000000000000000000000"
      --resume               resume withdrawal. use as --resume

Example1:
  ./intmax2-node tx withdrawal eth --amount 10 --recipient 0x32eD70FE0F69D6E915D27127fe6d0C016F20D2c2 --private-key 0x0000000000000000000000000000000000000000000000000000000000000002

Example2:
  ./intmax2-node tx withdrawal erc20 0x0000000000000000000000000000000000000001 --amount 10 --recipient 0x32eD70FE0F69D6E915D27127fe6d0C016F20D2c2 --private-key 0x0000000000000000000000000000000000000000000000000000000000000002

Example3:
  ./intmax2-node tx withdrawal erc721 0x0000000000000000000000000000000000000001 7 --amount 10 --recipient 0x32eD70FE0F69D6E915D27127fe6d0C016F20D2c2 --private-key 0x0000000000000000000000000000000000000000000000000000000000000002

Example4: Resume withdrawal transaction with recipient address.
  ./intmax2-node tx withdrawal --resume --recipient 0x32eD70FE0F69D6E915D27127fe6d0C016F20D2c2
```
### Command `./intmax2-node tx claim --help`
```
# ./intmax2-node tx claim --help
Send claim transaction

Usage:
  app tx claim [flags]

Flags:
  -h, --help                 help for claim
      --private-key string   specify user's Ethereum private key. use as --private-key "0x0000000000000000000000000000000000000000000000000000000000000000"
```

## Network
When a node starts, it tries to find and remember its external address in this order:
* If the `NETWORK_DOMAIN` and `NETWORK_PORT` environment variables are set, the node will publish the Block Builder address as `NETWORK_DOMAIN:NETWORK_PORT`.
* Before starting, the node attempts to determine its external IP address. To obtain an external address, an external STUN server is used (see settings for customizing the `STUN_SERVER` in `ENV Variables`)
* If the external IP address matches the local IP address, then the node will publish the Block Builder address as  the external IP and the port `HTTP_PORT` (see settings customizing the `HTTP (node)` in `ENV Variables`)
* If the external IP address is different from the local IP address, then the node will attempt a NAT Discover search for address `DOMAIN:PORT`, which was retrieved with help the `STUN` server. If successful, then the node will publish the Block Builder address as the external IP address and external port.
* If the previous steps are not successful, the node will offer to enter `NETWORK_DOMAIN` and `NETWORK_PORT` manually. In success, the node will publish the Block Builder address as `NETWORK_DOMAIN:NETWORK_PORT`.
* Once the Network is configured, the Network address will be committed to the blockchain.

## ENV Variables

| * | Variable                                                      | Default                                                            | Description                                                                                                                                |
|---|---------------------------------------------------------------|--------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
|   | **APP**                                                       |                                                                    |                                                                                                                                            |
|   | PRINT_CONFIG                                                  | false                                                              | displaing config info with start service                                                                                                   |
|   | CA_DOMAIN_NAME                                                | x.test.example.com                                                 | DNS.1 name of CA Root certificate                                                                                                          |
|   | PEM_PATH_CA_CERT                                              | scripts/x509/ca_cert.pem                                           | path to pem file with CA Root certificate                                                                                                  |
|   | PEM_PATH_SERV_CERT                                            | scripts/x509/server_cert.pem                                       | path to pem file with Server certificate                                                                                                   |
|   | PEM_PATH_SERV_KEY                                             | scripts/x509/server_key.pem                                        | path to pem file with Server key                                                                                                           |
|   | PEM_PATH_CA_CERT_CLIENT                                       | scripts/x509/client_ca_cert.pem                                    | path to pem file with CA Root Client certificate                                                                                           |
|   | PEM_PATH_CLIENT_CERT                                          | scripts/x509/client_cert.pem                                       | path to pem file with Client certificate                                                                                                   |
|   | PEM_PATH_CLIENT_KEY                                           | scripts/x509/client_key.pem                                        | path to pem file with Client key                                                                                                           |
|   | **PoW**                                                       |                                                                    |                                                                                                                                            |
|   | POW_DIFFICULTY                                                | 4000                                                               | the difficulty of proof-of-work                                                                                                            |
|   | POW_WORKERS                                                   | 2                                                                  | the number workers for compute PoW                                                                                                         |
|   | **BLOCKCHAIN**                                                |                                                                    |                                                                                                                                            |
|   | BLOCKCHAIN_SCROLL_NETWORK_CHAIN_ID                            |                                                                    | the Scroll blockchain network ID. Chain ID must be equal: ScrollSepolia = `534351`; Scroll = `534352`                                      |
|   | BLOCKCHAIN_SCROLL_MIN_BALANCE                                 | 100000000000000000                                                 | the Scroll blockchain balance minimal value for node start (min value equal or more then 0.1ETH)                                           |
|   | BLOCKCHAIN_SCROLL_STAKE_BALANCE                               | 100000000000000000                                                 | the Scroll blockchain balance value for stake with block builder update (min value equal or more then 0.1ETH)                              |
|   | BLOCKCHAIN_SCROLL_MESSENGER_L1_CONTRACT_ADDRESS               |                                                                    | the Scroll messagenger contract address on L1 Mainnet                                                                                      |
|   | BLOCKCHAIN_SCROLL_MESSENGER_L1_CONTRACT_DEPLOYED_BLOCK_NUMBER | 0                                                                  | the block number when the Scroll messagenger contract address on L1 Mainnet was deployed                                                   |
|   | BLOCKCHAIN_SCROLL_MESSENGER_L2_CONTRACT_ADDRESS               |                                                                    | the Scroll messagenger contract address on L2 Scroll                                                                                       |
| * | BLOCKCHAIN_BLOCK_BUILDER_REGISTRY_CONTRACT_ADDRESS            |                                                                    | the Block Builder Registry Contract address in the Scroll blockchain                                                                       |
| * | BLOCKCHAIN_ROLLUP_CONTRACT_ADDRESS                            |                                                                    | the Rollup Contract address in the Scroll blockchain                                                                                       |
|   | BLOCKCHAIN_ROLLUP_CONTRACT_DEPLOYED_BLOCK_NUMBER              | 0                                                                  | the block number when the Rollup contract was deployed                                                                                     |
| * | BLOCKCHAIN_LIQUIDITY_CONTRACT_ADDRESS                         |                                                                    | the Liquidity Contract address in the Mainnet                                                                                              |
|   | BLOCKCHAIN_LIQUIDITY_CONTRACT_DEPLOYED_BLOCK_NUMBER           | 0                                                                  | the block number when the Liquidity contract was deployed                                                                                  |
| * | BLOCKCHAIN_WITHDRAWAL_CONTRACT_ADDRESS                        |                                                                    | the Withdrawal Contract address in the Scroll blockchain                                                                                   |
|   | BLOCKCHAIN_ETHEREUM_NETWORK_CHAIN_ID                          |                                                                    | the Ethereum blockchain network ID. Chain ID must be equal: Sepolia = `11155111`; Ethereum = `1`                                           |
|   | BLOCKCHAIN_ETHEREUM_BUILDER_KEY_HEX                           |                                                                    | (pk) Ethereum builder private key                                                                                                          |
|   | BLOCKCHAIN_ETHEREUM_DEPOSIT_ANALYZER_PRIVATE_KEY_HEX          |                                                                    | (pk) Ethereum deposit analyzer private key                                                                                                 |
|   | BLOCKCHAIN_ETHEREUM_WITHDRAWAL_PRIVATE_KEY_HEX                |                                                                    | (pk) Ethereum withdrawal private key                                                                                                       |
|   | BLOCKCHAIN_ETHEREUM_MESSENEGER_MOCK_PRIVATE_KEY_HEX           |                                                                    | (pk) Ethereum messenger mock private key                                                                                                   |
|   | BLOCKCHAIN_MAX_COUNTER_OF_TRANSACTION                         | 128                                                                | max number of transactions in the block                                                                                                    |
|   | BLOCKCHAIN_DEPOSIT_ANALYZER_THRESHOLD                         | 10                                                                 | threshold for deposit analyzer                                                                                                             |
|   | BLOCKCHAIN_DEPOSIT_ANALYZER_MINUTES_THRESHOLD                 | 10                                                                 | minutes threshold for deposit analyzer                                                                                                     |
|   | **WALLET**                                                    |                                                                    |                                                                                                                                            |
|   | WALLET_PRIVATE_KEY_HEX                                        |                                                                    | (pk) private key for wallet address recognizing                                                                                            |
|   | WALLET_MNEMONIC_VALUE                                         |                                                                    | (mnemonic) mnemonic for recovery private key                                                                                               |
|   | WALLET_MNEMONIC_DERIVATION_PATH                               |                                                                    | (mnemonic) mnemonic password for recovery private key                                                                                      |
|   | WALLET_MNEMONIC_PASSWORD                                      |                                                                    | (mnemonic) mnemonic derivation path                                                                                                        |
|   | **LOG**                                                       |                                                                    |                                                                                                                                            |
|   | LOG_LEVEL                                                     | debug                                                              | log level                                                                                                                                  |
|   | IS_LOG_LINES                                                  | false                                                              | whether log line number of code where log func called or not                                                                               |
|   | LOG_JSON                                                      | false                                                              | when true prints each log message as separate JSON object                                                                                  |
|   | LOG_TIME_FORMAT                                               | 2006-01-02T15:04:05Z                                               | log time format in go time formatting style                                                                                                |
|   | **HTTP (node)**                                               |                                                                    |                                                                                                                                            |
|   | HTTP_CORS_ALLOW_ALL                                           | false                                                              | (node) allow all with cross-origin resource sharing                                                                                        |
|   | HTTP_CORS                                                     | *                                                                  | (node) cross-origin resource sharing                                                                                                       |
|   | HTTP_CORS_MAX_AGE                                             | 600                                                                | (node) default timeout in seconds of the cross-origin resource sharing                                                                     |
|   | HTTP_CORS_STATUS_CODE                                         | 204                                                                | (node) status code for the answer of the cross-origin resource sharing                                                                     |
|   | HTTP_HOST                                                     | 0.0.0.0                                                            | (node) host address for bind http external server                                                                                          |
|   | HTTP_PORT                                                     | 80                                                                 | (node) port for bind http external server                                                                                                  |
|   | HTTP_CORS_ALLOW_CREDENTIALS                                   | true                                                               | (node) allowed credentials                                                                                                                 |
|   | HTTP_CORS_ALLOW_METHODS                                       | OPTIONS                                                            | (node) allowed http methods                                                                                                                |
|   | HTTP_CORS_ALLOW_HEADS                                         | Accept;Authorization;Content-Type;X-CSRF-Token;X-User-Id;X-Api-Key | (node) allowed http heads                                                                                                                  |
|   | HTTP_CORS_EXPOSE_HEADS                                        |                                                                    | (node) exposed http methods                                                                                                                |
|   | HTTP_TLS_USE                                                  | false                                                              | (node) flag of turn off (false) or turn on (true) about use HTTPS                                                                          |
|   | COOKIE_SECURE                                                 | false                                                              | (node) flag of turn off (false) or turn on (true) the cookie secure (HTTP)                                                                 |
|   | COOKIE_DOMAIN                                                 |                                                                    | (node) name of domain for cookie                                                                                                           |
|   | COOKIE_SAME_SITE_STRICT_MODE                                  | false                                                              | (node) flag of turn off (false) or turn on (true) the cookie same site strict mode                                                         |
|   | COOKIE_FOR_AUTH_USE                                           | false                                                              | (node) flag of turn off (false) or turn on (true)  mode for the cookie use for authorization                                               |
|   | **GRPC (node)**                                               |                                                                    |                                                                                                                                            |
|   | GRPC_HOST                                                     | 0.0.0.0                                                            | (node) host address for bind gRPC external server                                                                                          |
|   | GRPC_PORT                                                     | 10000                                                              | (node) port for bind gRPC external server                                                                                                  |
|   | **SWAGGER (node)**                                            |                                                                    |                                                                                                                                            |
|   | SWAGGER_HOST_URL                                              | 127.0.0.1:8780                                                     | (node) host url for swagger-json connection                                                                                                |
|   | SWAGGER_BASE_PATH                                             | /                                                                  | (node) base path for swagger-json connection                                                                                               |
|   | **NETWORK**                                                   |                                                                    |                                                                                                                                            |
|   | NETWORK_DOMAIN                                                |                                                                    | `domain` or `ip-address` of external proxy-server for connections with node                                                                |
|   | NETWORK_PORT                                                  |                                                                    | `port` of external proxy-server for connections with node                                                                                  |
|   | NETWORK_HTTPS_USE                                             | false                                                              | flag of turn off (false) or turn on (true) about use HTTPS schema for external proxy-server for connections with node                      |
|   | **WITHDRAWAL_SERVICE**                                        |                                                                    |                                                                                                                                            |
| * | WITHDRAWAL_PROVER_URL                                         | http://localhost:8093                                              | API endpoint for verifying and processing withdrawal prover requests.                                                                      |
| * | WITHDRAWAL_GNARK_PROVER_URL                                   |                                                                    | API endpoint for verifying and processing Gnark prover requests.                                                                           |
|   | **API**                                                       |                                                                    |                                                                                                                                            |
|   | API_SCROLL_BRIDGE_URL                                         |                                                                    | API endpoint for verifying and processing scroll bridge requests.                                                                          |
| * | API_BLOCK_BUILDER_URL                                         |                                                                    | API endpoint for verifying and processing block builder requests.                                                                          |
| * | API_BLOCK_VALIDITY_PROVER_URL                                 |                                                                    | API endpoint for verifying and processing block validity prover requests.                                                                  |
| * | API_DATA_STORE_VAULT_URL                                      |                                                                    | API endpoint for verifying and processing data store vault requests.                                                                       |
| * | API_WITHDRAWAL_SERVER_URL                                     |                                                                    | API endpoint for verifying and processing withdrawal requests.                                                                             |
|   | **STUN SERVER**                                               |                                                                    |                                                                                                                                            |
| * | STUN_SERVER_NETWORK_TYPE                                      | udp6;udp4                                                          | network type for dial with stun server (separator equal `;`)                                                                               |
| * | STUN_SERVER_LIST                                              | stun.l.google.com:19302                                            | network address for dial with stun server (separator equal `;`)                                                                            |
|   | **GAS PRICE ORACLE**                                          |                                                                    |                                                                                                                                            |
|   | GAS_PRICE_ORACLE_EXTRA_FEE                                    | 0                                                                  | minimum value of extra fee that must be added to gasFee for transfer                                                                       |
|   | GAS_PRICE_ORACLE_DELIMITER                                    | 10                                                                 | minimum number of senders to which gasFee must be distributed for transfer                                                                 |
| * | GAS_PRICE_ORACLE_TIMEOUT                                      | 30s                                                                | timeout for updating gasFee from contract of the gas price oracle                                                                          |
|   | **WORKER**                                                    |                                                                    |                                                                                                                                            |
|   | WORKER_ID                                                     | pgx                                                                | id of worker                                                                                                                               |
|   | WORKER_PATH                                                   | /app/worker                                                        | dir of worker                                                                                                                              |
|   | WORKER_MAX_COUNTER                                            | 20                                                                 | max counter of worker                                                                                                                      |
|   | WORKER_PATH_CLEAN_IN_START                                    | 20                                                                 | flag of turn off (false) or turn on (true) the clean dir of worker                                                                         |
|   | WORKER_CURRENT_FILE_LIFETIME                                  | 1m                                                                 | timeout for create new current file of worker                                                                                              |
|   | WORKER_TIMEOUT_FOR_CHECK_CURRENT_FILE                         | 10s                                                                | timeout for check the status of current file of worker                                                                                     |
|   | WORKER_TIMEOUT_FOR_SIGNATURES_AVAILABLE_FILES                 | 15s                                                                | timeout for processing the transaction signature of current file of worker                                                                 |
|   | WORKER_MAX_COUNTER_OF_USERS                                   | 128                                                                | condition for create new of current file of worker                                                                                         |
|   | DEPOSIT_SYNCHRONIZER_ENABLE                                   | false                                                              | flag indicating whether to create an empty block when there is a deposit event that has not been reflected in the INTMAX block             |
|   | DEPOSIT_SYNCHRONIZER_TIMEOUT_FOR_EVENT_WATCHER                | 10s                                                                | interval at which Block Builder retrieves deposit events from on-chain                                                                     |
|   | **BALANCE VALIDITY PROVER**                                   |                                                                    |                                                                                                                                            |
|   | BALANCE_VALIDITY_PROVER_URL                                   | http://localhost:8092                                              | API endpoint for the balance validity prover                                                                                               |
|   | **BLOCK VALIDITY PROVER**                                     |                                                                    |                                                                                                                                            |
|   | BLOCK_VALIDITY_PROVER_URL                                     | http://localhost:8091                                              | API endpoint for the block validity prover                                                                                                 |
|   | BLOCK_VALIDITY_PROVER_EVENT_WATCHER_LIFETIME                  | 1m                                                                 | timeout for block validity prover event watcher                                                                                            |
|   | BLOCK_VALIDITY_PROVER_FETCH_PROOF_LIFETIME                    | 1m                                                                 | timeout for block validity prover fetch proof                                                                                              |
|   | BLOCK_VALIDITY_PROVER_MAX_VALUE_OF_INPUT_DEPOSITS_IN_REQUEST  | 10000                                                              | max value of the count of deposits for input request of webserver of the block validity prover                                             |
|   | BLOCK_VALIDITY_PROVER_MAX_VALUE_OF_INPUT_TX_ROOT_IN_REQUEST   | 10                                                                 | max value of the count of txs-root for input request of webserver of the block validity prover                                             |
|   | BLOCK_VALIDITY_PROVER_INVALID_TX_ROOT_IN_REQUEST              | 0xfe6fd7720cfd29168d72cff3db0a7a5ad31bd45195f9a9272bd367124a2989b3 | default for invalid value of the tx-root for input request of webserver of the block validity prover (the zero as tx-root always invalid)  |
|   | **SQL DB OF APP**                                             |                                                                    |                                                                                                                                            |
|   | SQL_DB_APP_DRIVER_NAME                                        | pgx                                                                | system driver name with sql driver of application (only, `pgx` of `postgres`)                                                              |
| * | SQL_DB_APP_DNS_CONNECTION                                     |                                                                    | connection string for connect with sql driver of application                                                                               |
|   | SQL_DB_APP_RECONNECT_TIMEOUT                                  | 1s                                                                 | timeout for database connection with sql driver of application                                                                             |
|   | SQL_DB_APP_OPEN_LIMIT                                         | 5                                                                  | maximum number of connections in the pool with sql driver of application                                                                   |
|   | SQL_DB_APP_IDLE_LIMIT                                         | 5                                                                  | the maximum number of connections in the idle with sql driver of application                                                               |
|   | SQL_DB_APP_CONN_LIFE                                          | 5m                                                                 | the maximum amount of time a connection may be reused with sql driver of application                                                       |
|   | **RECOMMIT OF SQL DB**                                        |                                                                    |                                                                                                                                            |
|   | SQL_DB_RECOMMIT_ATTEMPTS_NUMBER                               | 50                                                                 | attempts number tried of commits with transactions of sql driver                                                                           |
|   | SQL_DB_RECOMMIT_TIMEOUT                                       | 1s                                                                 | timeout of attempts number tried of commits with transactions of sql driver                                                                |
|   | **L2 BATCH INDEX**                                            |                                                                    |                                                                                                                                            |
| * | L2_BLOCK_NUMBER_TIMEOUT                                       | 10m                                                                | timeout for re-start of the L2 block number processing                                                                                     |
| * | L2_BATCH_INDEX_TIMEOUT                                        | 20m                                                                | timeout for re-start of the L2 batch index processing                                                                                      |
|   | **OPEN TELEMETRY**                                            |                                                                    |                                                                                                                                            |
|   | OPEN_TELEMETRY_ENABLE                                         | false                                                              | flag of turn off (false) or turn on (true) the OpenTelemetry                                                                               |
|   | OTEL_EXPORTER_OTLP_ENDPOINT                                   |                                                                    | external parameter (see official documentation about the OpenTelemetry; example: gRPC http://localhost:4317 or HTTP http://localhost:4318) |
|   | OTEL_EXPORTER_OTLP_COMPRESSION                                |                                                                    | external parameter (see official documentation about the OpenTelemetry; example = 'gzip')                                                  |

## Tests
For applies tests need copies `.env.example` as `.env` and runs command: `go test ./...`