ifndef ROOT_DIR
	export ROOT_DIR = $(shell git rev-parse --show-toplevel)
endif
export GIT_COMMIT_SHA = $(shell git rev-parse --short HEAD)

COMPOSE_RUN = docker compose run --rm --remove-orphans

start:
	@docker compose up --wait -d

stop: stop-api stop-quote
	@docker compose down --remove-orphans

clean: stop-api stop-quote
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

test-all: partition
	@$(COMPOSE_RUN) dev go test -tags integration ./...

partition:
	@bash scripts/setup_partitions.sh

build:
	-@docker rmi compass/app:$(GIT_COMMIT_SHA)
	@$(COMPOSE_RUN) dev go install ./...
	@docker build -t compass/app:$(GIT_COMMIT_SHA) -f dockerfiles/app.Dockerfile .

install: start migrate-up sqlc proto mock build

start-api:
	@$(COMPOSE_RUN) -d --name compass-api-service -p 50188:50188 -p 50189:50189 app api --grpc-addr :50188 --http-addr :50189

stop-api:
	-@docker stop compass-api-service

start-quote:
	@$(COMPOSE_RUN) -d --name compass-quote-service -p 50168:50168 app quote --grpc-addr :50168

stop-quote:
	-@docker stop compass-quote-service
