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
* (node) Serving gRPC server on http://0.0.0.0:10000
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
  completion         Generate the autocompletion script for the specified shell
  generate_account   Generate new Ethereum account
  help               Help about any command
  migrate            Execute migration
  mnemonic_account   Generate Ethereum account from mnemonic
  private_key_wallet Generate Ethereum wallet from private key
  run                run command
  stop               stop block builder command

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
### Command `./intmax2-node stop --help`
```
# ./intmax2-node stop --help
stop block builder command

Usage:
  app stop [flags]

Flags:
  -h, --help   help for stop
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
Generate new Ethereum account

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
Generate Ethereum account from mnemonic

Usage:
  app mnemonic_account [flags]

Flags:
      --derivation_path string     derivation_path flag. use as --derivation_path "m/44'/60'/0'/0/" (default "m/44'/60'/0'/0/")
  -h, --help                       help for mnemonic_account
      --key_number string          key_number flag. use as --key_number "0" (0 - parent account, 1...n - child accounts numbers)
      --mnemonic string            mnemonic flag. use as --mnemonic "mnemonic1 mnemonic2 ... mnemonic24"
      --mnemonic_password string   mnemonic_password flag. use as --mnemonic_password "pass"
```
### Command `./intmax2-node private_key_wallet --help`
```
# ./intmax2-node private_key_wallet --help
Generate Ethereum wallet from private key

Usage:
  app private_key_wallet [flags]

Flags:
  -h, --help                 help for private_key_wallet
      --private_key string   private_key flag. use as --private_key "__PRIVATE_KEY_IN_HEX_WITHOUT_0x__"
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

| * | Variable                                         | Default                                                            | Description                                                                                                                                |
|---|--------------------------------------------------|--------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
|   | **APP**                                          |                                                                    |                                                                                                                                            |
|   | PRINT_CONFIG                                     | false                                                              | displaing config info with start service                                                                                                   |
|   | CA_DOMAIN_NAME                                   | x.test.example.com                                                 | DNS.1 name of CA Root certificate                                                                                                          |
|   | PEM_PATH_CA_CERT                                 | scripts/x509/ca_cert.pem                                           | path to pem file with CA Root certificate                                                                                                  |
|   | PEM_PATH_SERV_CERT                               | scripts/x509/server_cert.pem                                       | path to pem file with Server certificate                                                                                                   |
|   | PEM_PATH_SERV_KEY                                | scripts/x509/server_key.pem                                        | path to pem file with Server key                                                                                                           |
|   | PEM_PATH_CA_CERT_CLIENT                          | scripts/x509/client_ca_cert.pem                                    | path to pem file with CA Root Client certificate                                                                                           |
|   | PEM_PATH_CLIENT_CERT                             | scripts/x509/client_cert.pem                                       | path to pem file with Client certificate                                                                                                   |
|   | PEM_PATH_CLIENT_KEY                              | scripts/x509/client_key.pem                                        | path to pem file with Client key                                                                                                           |
|   | **BLOCKCHAIN**                                   |                                                                    |                                                                                                                                            |
|   | BLOCKCHAIN_SCROLL_NETWORK_CHAIN_ID               |                                                                    | the Scroll blockchain network ID. Chain ID must be equal: ScrollSepolia = `534351`; Scroll = `534352`                                      |
|   | BLOCKCHAIN_SCROLL_MIN_BALANCE                    | 100000000000000000                                                 | the Scroll blockchain balance minimal value for node start (min value equal or more then 0.1ETH)                                           |
|   | BLOCKCHAIN_SCROLL_STAKE_BALANCE                  | 100000000000000000                                                 | the Scroll blockchain balance value for stake with block builder update (min value equal or more then 0.1ETH)                              |
| * | BLOCKCHAIN_ROLLUP_CONTRACT_ADDRESS               |                                                                    | the Rollup Contract address in the Scroll blockchain                                                                                       |
| * | BLOCKCHAIN_TEMPLATE_CONTRACT_ROLLUP_PATH         | templates/contracts/Rollup.json                                    | path to a file with information template for Rollup contract                                                                               |
|   | **WALLET**                                       |                                                                    |                                                                                                                                            |
|   | WALLET_PRIVATE_KEY_HEX                           |                                                                    | (pk) private key for wallet address recognizing                                                                                            |
|   | WALLET_MNEMONIC_VALUE                            |                                                                    | (mnemonic) mnemonic for recovery private key                                                                                               |
|   | WALLET_MNEMONIC_DERIVATION_PATH                  |                                                                    | (mnemonic) mnemonic password for recovery private key                                                                                      |
|   | WALLET_MNEMONIC_PASSWORD                         |                                                                    | (mnemonic) mnemonic derivation path                                                                                                        |
|   | **LOG**                                          |                                                                    |                                                                                                                                            |
|   | LOG_LEVEL                                        | debug                                                              | log level                                                                                                                                  |
|   | IS_LOG_LINES                                     | false                                                              | whether log line number of code where log func called or not                                                                               |
|   | LOG_JSON                                         | false                                                              | when true prints each log message as separate JSON object                                                                                  |
|   | LOG_TIME_FORMAT                                  | 2006-01-02T15:04:05Z                                               | log time format in go time formatting style                                                                                                |
|   | **HTTP (node)**                                  |                                                                    |                                                                                                                                            |
|   | HTTP_CORS_ALLOW_ALL                              | false                                                              | (node) allow all with cross-origin resource sharing                                                                                        |
|   | HTTP_CORS                                        | *                                                                  | (node) cross-origin resource sharing                                                                                                       |
|   | HTTP_CORS_MAX_AGE                                | 600                                                                | (node) default timeout in seconds of the cross-origin resource sharing                                                                     |
|   | HTTP_CORS_STATUS_CODE                            | 204                                                                | (node) status code for the answer of the cross-origin resource sharing                                                                     |
|   | HTTP_HOST                                        | 0.0.0.0                                                            | (node) host address for bind http external server                                                                                          |
|   | HTTP_PORT                                        | 80                                                                 | (node) port for bind http external server                                                                                                  |
|   | HTTP_CORS_ALLOW_CREDENTIALS                      | true                                                               | (node) allowed credentials                                                                                                                 |
|   | HTTP_CORS_ALLOW_METHODS                          | OPTIONS                                                            | (node) allowed http methods                                                                                                                |
|   | HTTP_CORS_ALLOW_HEADS                            | Accept;Authorization;Content-Type;X-CSRF-Token;X-User-Id;X-Api-Key | (node) allowed http heads                                                                                                                  |
|   | HTTP_CORS_EXPOSE_HEADS                           |                                                                    | (node) exposed http methods                                                                                                                |
|   | COOKIE_SECURE                                    | false                                                              | (node) flag of turn off (false) or turn on (true) the cookie secure (HTTP)                                                                 |
|   | COOKIE_DOMAIN                                    |                                                                    | (node) name of domain for cookie                                                                                                           |
|   | COOKIE_SAME_SITE_STRICT_MODE                     | false                                                              | (node) flag of turn off (false) or turn on (true) the cookie same site strict mode                                                         |
|   | COOKIE_FOR_AUTH_USE                              | false                                                              | (node) flag of turn off (false) or turn on (true)  mode for the cookie use for authorization                                               |
|   | **GRPC (node)**                                  |                                                                    |                                                                                                                                            |
|   | GRPC_HOST                                        | 0.0.0.0                                                            | (node) host address for bind gRPC external server                                                                                          |
|   | GRPC_PORT                                        | 10000                                                              | (node) port for bind gRPC external server                                                                                                  |
|   | **SWAGGER (node)**                               |                                                                    |                                                                                                                                            |
|   | SWAGGER_HOST_URL                                 | 127.0.0.1:8780                                                     | (node) host url for swagger-json connection                                                                                                |
|   | SWAGGER_BASE_PATH                                | /                                                                  | (node) base path for swagger-json connection                                                                                               |
|   | **NETWORK**                                      |                                                                    |                                                                                                                                            |
|   | NETWORK_DOMAIN                                   |                                                                    | `domain` or `ip-address` of external proxy-server for connections with node                                                                |
|   | NETWORK_PORT                                     |                                                                    | `port` of external proxy-server for connections with node                                                                                  |
|   | NETWORK_HTTPS_USE                                | false                                                              | flag of turn off (false) or turn on (true) about use HTTPS schema for external proxy-server for connections with node                      |
|   | **STUN SERVER**                                  |                                                                    |                                                                                                                                            |
| * | STUN_SERVER_NETWORK_TYPE                         | udp6;udp4                                                          | network type for dial with stun server (separator equal `;`)                                                                               |
| * | STUN_SERVER_LIST                                 | stun.l.google.com:19302                                            | network address for dial with stun server (separator equal `;`)                                                                            |
|   | **SQL DB OF APP**                                |                                                                    |                                                                                                                                            |
|   | SQL_DB_APP_DRIVER_NAME                           | pgx                                                                | system driver name with sql driver of application (only, `pgx` of `postgres`)                                                              |
| * | SQL_DB_APP_DNS_CONNECTION                        |                                                                    | connection string for connect with sql driver of application                                                                               |
|   | SQL_DB_APP_RECONNECT_TIMEOUT                     | 1s                                                                 | timeout for database connection with sql driver of application                                                                             |
|   | SQL_DB_APP_OPEN_LIMIT                            | 5                                                                  | maximum number of connections in the pool with sql driver of application                                                                   |
|   | SQL_DB_APP_IDLE_LIMIT                            | 5                                                                  | the maximum number of connections in the idle with sql driver of application                                                               |
|   | SQL_DB_APP_CONN_LIFE                             | 5m                                                                 | the maximum amount of time a connection may be reused with sql driver of application                                                       |
|   | **RECOMMIT OF SQL DB**                           |                                                                    |                                                                                                                                            |
|   | SQL_DB_RECOMMIT_ATTEMPTS_NUMBER                  | 50                                                                 | attempts number tried of commits with transactions of sql driver                                                                           |
|   | SQL_DB_RECOMMIT_TIMEOUT                          | 1s                                                                 | timeout of attempts number tried of commits with transactions of sql driver                                                                |
|   | **OPEN TELEMETRY**                               |                                                                    |                                                                                                                                            |
|   | OPEN_TELEMETRY_ENABLE                            | false                                                              | flag of turn off (false) or turn on (true) the OpenTelemetry                                                                               |
|   | OTEL_EXPORTER_OTLP_ENDPOINT                      |                                                                    | external parameter (see official documentation about the OpenTelemetry; example: gRPC http://localhost:4317 or HTTP http://localhost:4318) |
|   | OTEL_EXPORTER_OTLP_COMPRESSION                   |                                                                    | external parameter (see official documentation about the OpenTelemetry; example = 'gzip')                                                  |

## Tests
For applies tests need copies `.env.example` as `.env` and runs command: `go test ./...`