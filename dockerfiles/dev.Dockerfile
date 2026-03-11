FROM golangci/golangci-lint:v2.11.1-alpine

RUN apk add --no-cache curl docker-cli docker-cli-compose git jq make
