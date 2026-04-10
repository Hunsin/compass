FROM migrate/migrate:v4.19.1

COPY postgres/migrations /migrations

ENTRYPOINT ["migrate", "-path=/migrations"]
