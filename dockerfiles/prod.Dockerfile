FROM alpine:3.23.3

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app
COPY bin/compass /app/compass

ENTRYPOINT ["/app/compass"]
