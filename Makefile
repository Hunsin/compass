export GIT_COMMIT_SHA = $(shell git rev-parse --short HEAD)

COMPOSE_RUN = docker compose run --rm --remove-orphans

start:
	@docker compose up --wait -d

stop:
	@docker compose down --rmi local

go:
	@docker exec -it compass-dev sh

psql:
	@docker exec -it compass-postgres psql

# Database migration commands
migrate-up:
	@$(COMPOSE_RUN) migrate up

migrate-down:
	@$(COMPOSE_RUN) migrate down

migrate-version:
	@$(COMPOSE_RUN) migrate version