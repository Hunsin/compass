FROM golangci/golangci-lint:v2.8.0-alpine

RUN apk add --no-cache curl docker-cli docker-cli-compose git jq make protobuf

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1
