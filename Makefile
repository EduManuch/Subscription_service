# Устанавливаем цель по умолчанию. При вызове 'make' без аргументов будет показана справка.
.DEFAULT_GOAL := help

# Выносим повторяющиеся команды и имена в переменные для гибкости и уменьшения дублирования.
COMPOSE_FILE := docker-compose.yml
COMPOSE      := docker compose -f $(COMPOSE_FILE)

.PHONY: help up start down clean rebuild ps logs shell

# ---- HELP ----
help: ## показать цели
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*##' Makefile | awk 'BEGIN {FS=":"}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

up: ## собрать и запустить всё окружение
	$(COMPOSE) up -d --build --force-recreate
	@echo "\n🚀 App stack started"

start: ## запустить окружение без пересборки
	$(COMPOSE) start
	@echo "\n▶️ Services started"

down: ## остановить и удалить все контейнеры
	$(COMPOSE) down --remove-orphans -v
	@echo "\n🧹 All containers stopped and cleaned"

clean: ## очистить систему Docker
	$(COMPOSE) down --rmi local --remove-orphans
	@echo "\n🧽 Docker system cleaned."

rebuild: ## полная пересборка проекта
	$(MAKE) down
	$(MAKE) up
	@echo "\n♻️  Full environment rebuilt and started."

ps: ## список и статусы контейнеров
	$(COMPOSE) ps

logs: ## логи всех контейнеров
	$(COMPOSE) logs -f

shellp: ## shell в контейнер postgres
	docker exec -it postgres /bin/sh

psql: ## psql в postgres
	PGPASSWORD='password' psql -U postgres -h localhost -d sub_db

migrateup:
	migrate -path migrations -database "postgres://postgres:password@localhost:5432/sub_db?sslmode=disable" -verbose up

migratedown:
	migrate -path migrations -database "postgres://postgres:password@localhost:5432/sub_db?sslmode=disable" -verbose down


# ---- GO LOCAL ----
vet: ## запустить go vet
	go vet ./...

fmt: ## форматировать код
	go fmt ./...

run: ## запустить локально без Docker
	go run cmd/app/main.go