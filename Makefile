ifndef ROOT_DIR
	export ROOT_DIR = $(shell git rev-parse --show-toplevel)
endif
export GIT_COMMIT_SHA = $(shell git rev-parse --short HEAD)

COMPOSE_RUN = docker compose run --rm --remove-orphans

start:
	@docker compose up --wait -d

stop:
	@docker compose down --remove-orphans

clean:
	@docker compose down --remove-orphans --rmi local

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

sqlc:
	@$(COMPOSE_RUN) sqlc

proto:
	@docker exec compass-dev protoc \
		--go_out=. --go_opt=module=github.com/Hunsin/compass \
		--go-grpc_out=. --go-grpc_opt=module=github.com/Hunsin/compass \
		protocols/quote/quote.proto
