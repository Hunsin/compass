FROM golangci/golangci-lint:v2.10.1-alpine

RUN apk add --no-cache curl docker-cli docker-cli-compose git jq make
