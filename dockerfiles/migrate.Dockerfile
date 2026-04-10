FROM migrate/migrate:v4.19.1

RUN apk add --no-cache aws-cli

COPY postgres/migrations /migrations
COPY scripts/migrate-entrypoint.sh /usr/local/bin/migrate-entrypoint.sh
RUN chmod +x /usr/local/bin/migrate-entrypoint.sh

ENTRYPOINT ["migrate-entrypoint.sh"]
