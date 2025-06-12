.PHONY: migrate-up migrate-down migrate-status migrate-reset migrate-create db-start db-stop

# Database commands
db-start:
	docker-compose up -d postgres

db-stop:
	docker-compose down

db-logs:
	docker-compose logs -f postgres

# Migration commands
migrate-up:
	go run cmd/migrate/main.go postgres "$(DATABASE_URL)" up

migrate-down:
	go run cmd/migrate/main.go postgres "$(DATABASE_URL)" down

migrate-status:
	go run cmd/migrate/main.go postgres "$(DATABASE_URL)" status

migrate-reset:
	go run cmd/migrate/main.go postgres "$(DATABASE_URL)" reset

migrate-create:
	@read -p "Enter migration name: " name; \
	go run cmd/migrate/main.go postgres "$(DATABASE_URL)" create $$name sql

# Development
dev-setup: db-start
	@echo "Waiting for database to be ready..."
	@until docker-compose exec postgres pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB); do sleep 1; done
	$(MAKE) migrate-up

dev-reset: migrate-reset migrate-up

# Load environment
include .env
export $(shell sed 's/=.*//' .env)