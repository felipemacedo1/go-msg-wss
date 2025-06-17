# Makefile - comandos úteis para desenvolvimento

# Variáveis
DB_URL=postgres://postgres:admin@localhost:5432/msgwss?sslmode=disable
MIGRATIONS_DIR=./internal/store/pgstore/migrations

# Comandos

## gera código SQL com sqlc
sqlc-generate:
	@echo "🔄 Gerando código SQLC..."
	cd internal/store/pgstore && sqlc generate
	@echo "✅ Código SQLC gerado."

## reseta o banco (down -v + up + migrate)
reset-db:
	@echo "🚀 Resetando banco de dados..."
	docker-compose down -v
	docker-compose up -d db
	@echo "⏳ Aguardando Postgres iniciar..."
	sleep 10
	@echo "🗄️ Aplicando migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "✅ Banco de dados pronto."

## gera código e reseta banco (comando completo)
rebuild: sqlc-generate reset-db
	@echo "🎉 Rebuild completo finalizado!"

## sobe todos os serviços (db + pgadmin + outros)
up:
	docker-compose up -d

## derruba tudo
down:
	docker-compose down

## roda a aplicação
run:
	go run cmd/wsrs/main.go

## checa o status dos containers
ps:
	docker-compose ps

## roda testes (se tiver testes no projeto)
test:
	go test ./...

## formata o código
fmt:
	go fmt ./...