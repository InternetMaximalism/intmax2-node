ifneq (${GIT_USE},)
ifeq ($(shell git tag --contains HEAD),)
  VERSION := $(shell git rev-parse --short HEAD)
else
  VERSION := $(shell git tag --contains HEAD)
endif
endif

ifneq ($(goproxy),)
  re_build_arg := --build-arg goproxy="$(goproxy)"
endif

ifeq ($(shell uname -s),Darwin)
	SED_COMMAND := gsed
else
	SED_COMMAND := sed
endif

BUILDNAME := intmax2-node
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOLDFLAGS += -X intmax2-node/configs/buildvars.Version=$(VERSION)
GOLDFLAGS += -X intmax2-node/configs/buildvars.BuildTime=$(BUILDTIME)
GOLDFLAGS += -X intmax2-node/configs/buildvars.BuildName=$(BUILDNAME)
GOFLAGS = -ldflags "$(GOLDFLAGS)"

.DEFAULT_GOAL := default

.PHONY: default
default: gen build #lint build

.PHONY: build
build:
	go build -v -o $(BUILDNAME) $(GOFLAGS) ./cmd/

.PHONY: gen
gen: format-proto
	buf generate -v --debug --timeout=2m --template api/proto/service/buf.gen.yaml api/proto/service
	buf generate -v --debug --timeout=2m --template api/proto/service/buf.gen.tagger.yaml api/proto/service
	go generate -v ./...
	cp -rf docs/swagger/node third_party/OpenAPI
ifneq (${SWAGGER_USE},)
# generic values
ifneq (${SWAGGER_BUILD_MODE},)
	$(SED_COMMAND) -i "s/SWAGGER_VERSION/$(VERSION)/g" third_party/OpenAPI/node/apidocs.swagger.json
else
	$(SED_COMMAND) -i "s/SWAGGER_VERSION/v0.0.0/g" third_party/OpenAPI/node/apidocs.swagger.json
endif
ifneq (${SWAGGER_HOST_URL},)
	$(SED_COMMAND) -i "s/SWAGGER_HOST_URL/${SWAGGER_HOST_URL}/g" third_party/OpenAPI/node/apidocs.swagger.json
else
	$(SED_COMMAND) -i "s/SWAGGER_HOST_URL//g" third_party/OpenAPI/node/apidocs.swagger.json
endif
ifneq (${SWAGGER_BASE_PATH},)
	$(SED_COMMAND) -i "s/SWAGGER_BASE_PATH/${SWAGGER_BASE_PATH}/g" third_party/OpenAPI/node/apidocs.swagger.json
else
	$(SED_COMMAND) -i "s/SWAGGER_BASE_PATH/\//g" third_party/OpenAPI/node/apidocs.swagger.json
endif
endif

.PHONY: format-proto
format-proto: ## format all protos
	clang-format -i api/proto/service/node/v1/node.proto

.PHONY: tools
tools:
	go install -v go.uber.org/mock/mockgen@latest
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.0
	go install -v github.com/bufbuild/buf/cmd/buf@v1.34.0
	go install -v github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.16.1
	go install -v github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.16.1
	go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0
	go install -v google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
	go install -v github.com/srikrsna/protoc-gen-gotag@v1.0.1

.PHONY: run
run: gen ## starting application and dependency services
# translate `SWAGGER_USE=true GIT_USE=true CMD_RUN="run" make run` => ./intmax2-node run
	go run $(GOFLAGS) ./cmd ${CMD_RUN}

.PHONY: up
up: ## starting application and dependency services
	cp -f build/env.docker.node.example build/env.docker.node
	docker compose -f build/docker-compose.yml up

.PHONY: build-up
build-up: down ## rebuilding containers and starting application and dependency services
	cp -f build/env.docker.node.example build/env.docker.node
	docker compose -f build/docker-compose.yml build $(re_build_arg)
	docker compose -f build/docker-compose.yml up

.PHONY: start-build-up
start-build-up: down ## rebuilding containers and starting application and dependency services
	cp -f build/env.docker.node.example build/env.docker.node
	docker compose -f build/docker-compose.yml up -d intmax2-node-postgres
	docker compose -f build/docker-compose.yml up -d intmax2-node-ot-collector


.PHONY: down
down:
	cp -f build/env.docker.node.example build/env.docker.node
	docker compose -f build/docker-compose.yml down
	rm -f build/env.docker.node

.PHONY: clean-all
clean-all: down
	rm -f build/env.docker.node
	rm -rf build/sql_dbs/intmax2-node-postgres

.PHONY: lint
lint:
	buf lint api/proto/service
	golangci-lint run --timeout=10m --fix ./...