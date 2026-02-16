export GIT_COMMIT_SHA = $(shell git rev-parse --short HEAD)

start:
	@docker compose up -d

stop:
	@docker compose down --rmi local

go:
	@docker exec -it compass-dev sh
