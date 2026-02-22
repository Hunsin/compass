FROM golangci/golangci-lint:v2.10.1-alpine

RUN apk add --no-cache curl docker-cli docker-cli-compose git jq make

# Manually install protoc because the latest version (31.1) in Alpine 3.23 doesn't
# support edition 2024
ARG PROTOC_VERSION=33.5
ARG PROTOC_RELEASES=https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION
RUN if [ $(uname -m) = "aarch64" ]; then \
        export PROTOP_ZIP=protoc-$PROTOC_VERSION-linux-aarch_64.zip; \
    elif [ $(uname -m) = "x86_64" ]; then \
        export PROTOP_ZIP=protoc-$PROTOC_VERSION-linux-x86_64.zip; \
    else \
        echo "unsupport target $(uname -m)"; exit 1;\
    fi; \
    curl -LO $PROTOC_RELEASES/$PROTOP_ZIP; \
    unzip $PROTOP_ZIP -d /usr/local; \
    rm $PROTOP_ZIP

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11; \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1
