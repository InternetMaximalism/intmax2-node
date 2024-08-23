FROM golang:1.21.9-alpine3.19 AS intmax2-build-tools
RUN apk add make git clang-extra-tools
WORKDIR /src
#ARG goproxy=direct
#ENV GOPROXY=$goproxy
ENV CGO_ENABLED=0
RUN go install -v go.uber.org/mock/mockgen@latest
RUN go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.0
RUN go install -v github.com/bufbuild/buf/cmd/buf@v1.34.0
RUN go install -v github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.16.1
RUN go install -v github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.16.1
RUN go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0
RUN go install -v google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
RUN go install -v github.com/srikrsna/protoc-gen-gotag@v1.0.1

FROM intmax2-build-tools AS intmax2-build-env
ENV GIT_USE=true
ENV SWAGGER_USE=true
ENV SWAGGER_BUILD_MODE=true
ARG SWAGGER_HOST_URL
ENV SWAGGER_HOST_URL=$SWAGGER_HOST_URL
ARG SWAGGER_BASE_PATH
ENV SWAGGER_BASE_PATH=$SWAGGER_BASE_PATH
WORKDIR /src
ADD . .
ARG goproxy=direct
ENV GOPROXY=$goproxy
ENV CGO_ENABLED=0
RUN make

FROM alpine:3.19 AS intmax2-run-env
COPY --from=intmax2-build-env /src/intmax2-node /app/intmax2-node
WORKDIR /app
