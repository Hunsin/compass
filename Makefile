ifndef ROOT_DIR
	export ROOT_DIR = $(shell git rev-parse --show-toplevel)
endif
export GIT_COMMIT_SHA = $(shell git rev-parse --short HEAD)

COMPOSE_RUN = docker compose run --rm --remove-orphans

start:
	@docker compose up --wait -d

stop: stop-quote
	@docker compose down --remove-orphans

clean: stop-quote
	-@docker compose down --remove-orphans --rmi local --volumes
	-@docker volume rm compass-app

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
	@$(COMPOSE_RUN) buf

mock:
	@$(COMPOSE_RUN) mockery

lint:
	@$(COMPOSE_RUN) dev golangci-lint run

test:
	@$(COMPOSE_RUN) dev go test ./...

test-all: partition-default
	@$(COMPOSE_RUN) dev go test -tags integration ./...

build:
	@$(COMPOSE_RUN) -e CGO_ENABLED=0 dev go build -o bin/compass ./cmd/compass

partition-default:
	@bash scripts/setup_partitions.sh

install: start migrate-up sqlc proto mock
	@$(COMPOSE_RUN) dev go install ./...

start-quote:
	@$(COMPOSE_RUN) -d --name compass-quote-service -p 50168:50168 app quote --listen-addr :50168

stop-quote:
	-@docker stop compass-quote-service
