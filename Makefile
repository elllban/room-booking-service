.PHONY: up down seed test cover

# Запуск проекта
up:
	docker-compose up --build

# Остановка проекта
down:
	docker-compose down -v

# Наполнение БД тестовыми данными
seed:
	docker-compose exec app go run scripts/seed.go

# Запуск тестов
test:
	go test ./... -v

# Покрытие тестами
cover:
	go test ./... -cover